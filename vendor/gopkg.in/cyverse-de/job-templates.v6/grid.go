package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v4"
)

// GridJobSubmissionBuilder is responsible for writing out the iplant.cmd,
// config, and job files in the directory specififed by dirPath, but only for
// job submissions to our local HTCondor cluster.
type GridJobSubmissionBuilder struct {
	cfg *viper.Viper
}

// Build is where the the iplant.cmd, config, and job files are actually written
// out for submissions to the local HTCondor cluster.
func (b GridJobSubmissionBuilder) Build(submission *model.Job, dirPath string) (string, error) {
	var err error

	templateFields := OtherTemplateFields{
		PathListHeader:       b.cfg.GetString("path_list.file_identifier"),
		TicketPathListHeader: b.cfg.GetString("tickets_path_list.file_identifier"),
	}
	templateModel := TemplatesModel{
		submission,
		templateFields,
	}

	submission.OutputTicketFile, err = generateOutputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

	submission.InputTicketsFile, err = generateInputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

	submission.InputPathListFile, err = generateInputPathList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

	// Generate the submission file.
	submitFilePath, err := generateFile(dirPath, "iplant.cmd", gridSubmissionTemplate, submission)
	if err != nil {
		return "", err
	}

	// Generate the job configuration file.
	_, err = generateFile(dirPath, "config", gridJobConfigTemplate, b.cfg)
	if err != nil {
		return "", err
	}

	// Write the job submission to a JSON file.
	_, err = generateJSON(dirPath, "job", submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}

func newGridJobSubmissionBuilder(cfg *viper.Viper) JobSubmissionBuilder {
	return GridJobSubmissionBuilder{cfg: cfg}
}
