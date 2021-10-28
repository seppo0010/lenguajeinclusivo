package shared

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetSha1(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ReadSecret(name string) (string, error) {
	body, err := ioutil.ReadFile(fmt.Sprintf("/run/secrets/%v", name))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"name":  name,
		}).Fatal("failed to read secret")
	}
	return strings.TrimSpace(string(body)), nil
}
