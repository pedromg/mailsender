package main

import (
	"fmt"
	"testing"
)

const (
	jsonFile = "./test/mailsender.json"
	l100     = "...................................................................................................."
)

var (
	l500 = fmt.Sprintf("%s%s%s%s%s", l100, l100, l100, l100, l100)
)

func TestConfigValidate(t *testing.T) {
	type configsCase struct {
		C    *Configs
		Pass bool
		Errs []string
	}

	cases := []configsCase{
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				ServerName:   "LinuxBox",
				AppName:      "App",
				Subject:      "subject line",
				Body:         "body text",
			},
			true,
			nil,
		},

		configsCase{
			&Configs{
				Configs:      "./test/no_file",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender_too_big.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "bad email",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "bad email",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				ServerName:   l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				AppName:      l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				Subject:      l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				Body:         l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				Message:      l500,
			},
			false,
			nil,
		},
	}

	for nr, c := range cases {
		if err := c.C.validate(); (err != nil) == c.Pass {
			t.Errorf("Case %d Pass should be %t", nr, c.Pass)
		}
	}
}

func TestFromJSON(t *testing.T) {
	type Case struct {
		Data string
		Pass bool
	}

	cases := []Case{
		{
			"{ }",
			true,
		},
		{
			"{ \"\"}",
			false,
		},
		{
			"{ \"neglected_well_formed_garbage\": \"\"}",
			true,
		},
		{
			"{ \"server_name\": \"server\", \"app_name\": \"app\", \"secure\": true }",
			true,
		},
	}

	for nr, c := range cases {
		configs := &Configs{}
		if err := configs.fromJSON([]byte(c.Data)); (err != nil) == c.Pass {
			t.Errorf("Case %d Pass should be %t", nr, c.Pass)
		}

		// in case of the JSON parse passes, check for some other conditions
		if c.Pass {
			// t.Logf(">>> Subject: %s", configs.Subject)
			if configs.Subject == "" {
				t.Errorf("Case %d error: Subject should not be empty: %s", nr, configs.Subject)
			}
			// t.Logf(">>> Body: %s", configs.Body)
			if configs.Body == "" {
				t.Errorf("Case %d error: Body should not be empty: %s", nr, configs.Body)
			}
		}

	}
}

func TestLoadConfigs(t *testing.T) {
	type Case struct {
		Filename string
		Pass     bool
	}

	cases := []Case{
		{
			"",
			false,
		},
		{
			"./test",
			false,
		},
		{
			"./test/mailsender_too_big.json",
			false,
		},
		{
			"./test/mailsender.json",
			true,
		},
	}

	for nr, c := range cases {
		configs := &Configs{}
		if err := configs.load(c.Filename); (err != nil) == c.Pass {
			t.Errorf("Case %d Pass should be %t", nr, c.Pass)
		}

		// in case of the JSON parse passes, check for some other conditions
		if c.Pass {
			// t.Logf(">>> Subject: %s", configs.Subject)
			if configs.Subject == "" {
				t.Errorf("Case %d error: Subject should not be empty: %s", nr, configs.Subject)
			}
			// t.Logf(">>> Body: %s", configs.Body)
			if configs.Body == "" {
				t.Errorf("Case %d error: Body should not be empty: %s", nr, configs.Body)
			}
		}
	}
}
