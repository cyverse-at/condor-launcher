package jobs

import (
	"fmt"
	"log"
	"text/template"

	"github.com/pkg/errors"
)

var (
	// gridSubmissionTemplate is a *template.Template for the HTCondor submission file
	gridSubmissionTemplate *template.Template

	// gridJobConfigTemplate is the *template.Template for the job definition JSON
	gridJobConfigTemplate *template.Template
)

// SubmissionTemplateText is the text of the template for the HTCondor
// submission file.
const gridSubmissionTemplateText = `universe = grid
grid_resource = condor SAURON1.pers.ad.uni-graz.at SAURON1.pers.ad.uni-graz.at

+remote_universe = 9
+remote_gridresource = "sge"
+remote_ShouldTransferFiles = "YES"
+remote_WhenToTransferOutput = "ON_EXIT"
+remote_queue = "sge"
+remote_batchqueue = "all.q"

executable = /software/cyverse/entrypoint
transfer_executable = False

+remote_BatchExtraSubmitArgs = "#$-l h_vmem={{ sgeBytes .MemoryRequest .CPURequest 8589934592 }}\n{{- if .CPURequest }}#$-pe smp {{ .CPURequest }}\n{{ end }}{{- if .DiskRequest }}#$-l tmpspace={{ sgeBytes .DiskRequest 0 0 }}\n{{ end }}"

arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log

accounting_group = {{if .Group}}{{.Group}}{{else}}de{{end}}
accounting_group_user = {{.Submitter}}
+IpcUuid = "{{.InvocationID}}"
+IpcJobId = "generated_script"
+IpcUsername = "{{.Submitter}}"
+IpcUserGroups = {{.FormatUserGroups}}
concurrency_limits = {{.UserIDForSubmission}}

{{with $x := index .Steps 0}}+IpcExe = "{{$x.Component.Name}}"{{end}}
{{with $x := index .Steps 0}}+IpcExePath = "{{$x.Component.Location}}"{{end}}
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
{{- if .OutputTicketFile -}}
,{{.OutputTicketFile}}
{{- end}}
{{- if .InputTicketsFile -}}
,{{.InputTicketsFile}}
{{- end}}
{{- if .InputPathListFile -}}
,{{.InputPathListFile}}
{{- end}}
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`

// JobConfigTemplateText is the text of the template for the HTCondor submission
// file.
const gridJobConfigTemplateText = `
amqp:
    uri: {{.GetString "amqp.uri"}}
    exchange:
        name: {{.GetString "amqp.exchange.name"}}
        type: {{.GetString "amqp.exchange.type"}}
irods:
    base: "{{.GetString "irods.base"}}"
porklock:
    image: "{{.GetString "porklock.image"}}"
    tag: "{{.GetString "porklock.tag"}}"
condor:
    filter_files: "{{.GetString "condor.filter_files"}}"
vault:
    token: "{{.GetString "vault.child_token.token"}}"
    url: "{{.GetString "vault.url"}}"
`

// SGEBytes formats a number of bytes to a condor format (rounding up to the nearest KiB until it's at least 1MiB, then rounding up to the nearest MiB)
func SGEBytes(bytes int64, cores float32, minimum int64) string {
	if cores > 1.0 {
		bytes = bytes / int64(cores)
	}

	if minimum > 0 && bytes < minimum {
		bytes = minimum
	}

	if bytes < bytesPerKiB {
		return "1K"
	}

	if bytes < bytesPerMiB {
		kb := bytes / bytesPerKiB
		if bytes%bytesPerKiB > 0 {
			kb = kb + 1
		}
		return fmt.Sprintf("%dK", kb)
	}

	mb := bytes / bytesPerMiB
	if bytes%bytesPerMiB > 0 {
		mb = mb + 1
	}

	return fmt.Sprintf("%dM", mb)
}

func init() {
	var err error

	funcMap := template.FuncMap{
		"sgeBytes": SGEBytes,
	}

	gridSubmissionTemplate, err = template.New("condor_submit").Funcs(funcMap).Parse(gridSubmissionTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse submission template text"))
	}
	gridJobConfigTemplate, err = template.New("job_config").Parse(gridJobConfigTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse job config template text"))
	}

	inputPathListTemplate, err = template.New("input_path_list").Parse(inputPathListTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse input path list template text"))
	}

	inputTicketListTemplate, err = template.New("input_tickets").Parse(inputTicketListTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse input tickets template text"))
	}
	outputTicketListTemplate, err = template.New("output_ticket").Parse(outputTicketListTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse output ticket template text"))
	}
}
