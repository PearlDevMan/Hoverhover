package hoverfly

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

func (d *DBClient) Import(uri string) error {

	// assuming file URI is URL:
	if IsURL(uri) {
		log.WithFields(log.Fields{
			"isURL":      IsURL(uri),
			"importFrom": uri,
		}).Info("URL")
		return d.ImportFromUrl(uri)
	}
	// assuming file URI is disk location
	ext := path.Ext(uri)
	if ext != ".json" {
		return fmt.Errorf("Failed to import payloads, only JSON files are acceppted. Given file: %s", uri)
	} else {
		// checking whether it exists
		exists, err := exists(uri)
		if err != nil {
			return fmt.Errorf("Failed to import payloads from %s. Got error: %s", uri, err.Error())
		}
		if exists {
			// file is JSON and it exist
			return d.ImportFromDisk(uri)
		} else {
			return fmt.Errorf("Failed to import payloads, given file '%s' does not exist", uri)
		}
	}
}

const URL string = `^((ftp|https?):\/\/)(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(([a-zA-Z0-9]+([-\.][a-zA-Z0-9]+)*)|((www\.)?))?(([a-z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-z\x{00a1}-\x{ffff}]{2,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`

var rxURL = regexp.MustCompile(URL)

func IsURL(str string) bool {
	if str == "" || len(str) >= 2083 || len(str) <= 3 || strings.HasPrefix(str, ".") {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return rxURL.MatchString(str)

}
func (d *DBClient) ImportPayloads(payloads []Payload) error {
	if len(payloads) > 0 {
		success := 0
		failed := 0
		for _, pl := range payloads {
			// recalculating request hash and storing it in database
			r := RequestContainer{Details: pl.Request}
			key := r.Hash()

			// regenerating key
			pl.ID = key

			bts, err := pl.Encode()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("Failed to encode payload")
				failed += 1
			} else {
				// hook
				var en Entry
				en.ActionType = ActionTypeRequestCaptured
				en.Message = "imported"
				en.Time = time.Now()
				en.Data = bts

				if err := d.Hooks.Fire(ActionTypeRequestCaptured, &en); err != nil {
					log.WithFields(log.Fields{
						"error":      err.Error(),
						"message":    en.Message,
						"actionType": ActionTypeRequestCaptured,
					}).Error("failed to fire hook")
				}

				d.Cache.Set([]byte(key), bts)
				if err == nil {
					success += 1
				} else {
					failed += 1
				}
			}
		}
		log.WithFields(log.Fields{
			"total":      len(payloads),
			"successful": success,
			"failed":     failed,
		}).Info("payloads imported")
		return nil
	} else {
		return fmt.Errorf("Bad request. Nothing to import!")
	}

}
