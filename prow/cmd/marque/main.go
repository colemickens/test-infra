/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"k8s.io/test-infra/prow/kube"
)

const (
	// Time before generating the first cert, for safety.
	generateWaitTime = 15 * time.Second
	// Time between renewals.
	renewTime = 12 * time.Hour
)

var (
	email      string
	domains    []string
	secretName string
)

func init() {
	email = os.Getenv("EMAIL")
	if email == "" {
		email = "spxtr@google.com"
	}

	domainsRaw := os.Getenv("DOMAINS")
	if domainsRaw != "" {
		domains = strings.Split(domainsRaw, ",")
	} else {
		domains = []string{"prow.k8s.io", "prow.kubernetes.io"}
	}

	secretName = os.Getenv("SECRET_NAME")
	if secretName == "" {
		secretName = "prow-k8s-cert"
	}
}

func main() {
	kc, err := kube.NewClientInCluster("default")
	if err != nil {
		logrus.WithError(err).Fatal("Error getting kube client.")
	}

	root, err := ioutil.TempDir("", "certbot")
	if err != nil {
		logrus.WithError(err).Fatal("Could not create temp dir.")
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/cole", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hello")) })
	http.Handle("/.well-known/", http.FileServer(http.Dir(root)))
	go func() { logrus.WithError(http.ListenAndServe(":http", nil)).Fatal("Server returned.") }()

	logrus.Infof("Sleeping for %v before generating cert.", generateWaitTime)
	for range time.Tick(generateWaitTime) {
		if err := generate(root); err != nil {
			logrus.WithError(err).Warning("Error renewing cert.")
			continue
		}
		if err := replaceSecret(kc); err != nil {
			logrus.WithError(err).Warning("Error updating secrets.")
			continue
		}
		break
	}

	for range time.Tick(renewTime) {
		if err := renew(); err != nil {
			logrus.WithError(err).Warning("Error renewing cert.")
		}
		if err := replaceSecret(kc); err != nil {
			logrus.WithError(err).Warning("Error updating secrets.")
		}
	}
}

func generate(root string) error {
	args := []string{
		"certonly",
		"--agree-tos",
		"--email", email,
		"--non-interactive",
		"--webroot",
		"-vvv",
		"-w", root,
	}
	for _, domain := range domains {
		args = append(args, "-d", domain)
	}

	logrus.Infof("Running: certbot %s", strings.Join(args, " "))
	cmd := exec.Command("certbot", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certbot error: %v output: %s", err, string(output))
	}
	logrus.Infof("Finished executing certbot.")
	return nil
}

func renew() error {
	logrus.Info("Running: certbot renew")
	cmd := exec.Command("certbot", "renew")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certbot error: %v output: %s", err, string(output))
	}
	return nil
}

func replaceSecret(c *kube.Client) error {
	key, err := ioutil.ReadFile("/etc/letsencrypt/live/" + domains[0] + "/privkey.pem")
	if err != nil {
		return fmt.Errorf("could not read privkey: %v", err)
	}
	cert, err := ioutil.ReadFile("/etc/letsencrypt/live/" + domains[0] + "/fullchain.pem")
	if err != nil {
		return fmt.Errorf("could not read fullchain: %v", err)
	}

	s := kube.Secret{
		Metadata: kube.ObjectMeta{
			Name: secretName,
		},
		Data: map[string]string{
			"tls.crt": base64.StdEncoding.EncodeToString(cert),
			"tls.key": base64.StdEncoding.EncodeToString(key),
		},
	}

	fmt.Printf("%v", s)

	if err := c.ReplaceSecret(secretName, s); err != nil {
		return fmt.Errorf("could not replace secret: %v", err)
	}
	return nil
}
