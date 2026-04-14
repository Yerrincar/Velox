package immich

import (
	"fmt"
	"strconv"
)

func ImmichUpload(folder, instanceURL, apiKey string) string {
	return fmt.Sprintf(
		"IMMICH_INSTANCE_URL=%s IMMICH_API_KEY=%s immich upload --recursive -c 8 --delete %s",
		strconv.Quote(instanceURL),
		strconv.Quote(apiKey),
		strconv.Quote(folder),
	)
}
