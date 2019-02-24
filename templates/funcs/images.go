package funcs

import (
	"strconv"
	"strings"

	"github.com/ernestoalejo/aeimagesflags"

	"libs.altipla.consulting/errors"
)

func Thumbnail(servingURL string, strFlags string) (string, error) {
	if servingURL == "" || strFlags == "" {
		return "", nil
	}

	flags := aeimagesflags.Flags{
		ExpiresDays: 365,
	}
	for _, part := range strings.Split(strFlags, ";") {
		strFlag := strings.Split(part, "=")
		if len(strFlag) != 2 {
			return "", errors.Errorf("all flags should be in the form key=value")
		}

		switch strings.TrimSpace(strFlag[0]) {
		case "width":
			n, err := strconv.ParseUint(strFlag[1], 10, 64)
			if err != nil {
				return "", errors.Wrapf(err, "cannot parse width flag")
			}
			flags.Width = n

		case "height":
			n, err := strconv.ParseUint(strFlag[1], 10, 64)
			if err != nil {
				return "", errors.Wrapf(err, "cannot parse height flag")
			}
			flags.Height = n

		case "square-crop":
			flags.SquareCrop = (strFlag[1] == "true")

		case "smart-square-crop":
			flags.SmartSquareCrop = (strFlag[1] == "true")

		case "original":
			flags.Original = (strFlag[1] == "true")

		case "size":
			n, err := strconv.ParseUint(strFlag[1], 10, 64)
			if err != nil {
				return "", errors.Wrapf(err, "cannot parse size flag")
			}
			flags.Size = n

		default:
			return "", errors.Errorf("unknown image flag: %s", strFlag[0])
		}
	}

	servingURL = strings.Replace(servingURL, "http://", "https://", 1)
	return aeimagesflags.Apply(servingURL, flags), nil
}
