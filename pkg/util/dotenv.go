package util

import (
	"os"
	"path"

	"github.com/joho/godotenv"
)

// Load all .env in project root, return first error that occurs.
func DotenvLoad(filenames ...string) error {
	for _, filename := range filenames {
		file := path.Join(DirName(), "../../", filename)
		if _, err := os.Stat(file); err == nil {
			err := godotenv.Load(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
