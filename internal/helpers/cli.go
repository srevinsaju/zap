package helpers

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/srevinsaju/zap/exceptions"
	"github.com/srevinsaju/zap/types"
)

type InteractiveSurveyOptions struct {
	Classifier string
	Array      []string
	Default    string
	Options    types.InstallOptions
}

func InteractiveSurvey(options InteractiveSurveyOptions) (string, error) {
	userResponse := ""
	if len(options.Array) == 0 {
		logger.Debugf("⚠️ Couldn't find any %ss", options.Classifier)
		return "", exceptions.NoReleaseFoundError
	} else if len(options.Array) == 1 {
		// directly select that release coz. there is only one release
		logger.Debugf("Found one %s. Selecting that as default", options.Classifier)
		userResponse = options.Array[0]
	} else if options.Options.Silent {
		// user has requested silence
		// we should not prompt the user and ask for selecting an option from this
		return "", exceptions.SilenceRequestedError
	} else {
		// there are a lot of items in the release, hmm...
		logger.Debugf("Preparing survey for %s selection", options.Classifier)
		prompt := &survey.Select{
			Message: fmt.Sprintf("Choose a %s", options.Classifier),
			Options: options.Array,
			Default: options.Array[0],
		}
		err := survey.AskOne(prompt, &userResponse)
		if err != nil {
			return "", err
		}
	}
	return userResponse, nil
}
