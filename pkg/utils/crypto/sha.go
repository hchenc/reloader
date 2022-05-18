package crypto

import (
	"crypto/sha1"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
)

// GenerateSHA generates SHA from string
func GenerateSHA(data string) string {
	hasher := sha1.New()
	if _, err := io.WriteString(hasher, data); err != nil {
		logrus.Errorf("Unable to write data in hash writer %v", err)
	}
	sha := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha)
}
