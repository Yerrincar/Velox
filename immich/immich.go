package immich

func ImmichUpload(folder string) string {
	return "immich upload --recursive -c 8 --delete " + folder
}
