package main

import (
	"fmt"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

var tcRcpts = []struct {
	name  string
	to    []string
	expTo string
}{
	// -t email1@sensu.com
	{"single", []string{"email1@example.com"}, "To: email1@example.com"},
	// -t email1@sensu.com,email2@sensu.com
	{"single_comma", []string{"email1@example.com,email2@example.com"}, "To: email1@example.com,email2@example.com"},
	// -t "email1@sensu.com, email2@sensu.com"
	{"single_comma_space", []string{"email1@example.com, email2@example.com"}, "To: email1@example.com,email2@example.com"},
	// -t email1@sensu.com -t email2@sensu.com
	{"multiple_flag", []string{"email1@example.com", "email2@example.com"}, "To: email1@example.com,email2@example.com"},
	// -t " email1@example.com\r\n, email2@example.com" -t email3@sensu.com
	// note invalid line endings removed, and order is changed
	{"multiple_flag_comma", []string{" email1@example.com\r\n, email2@example.com", "email3@example.com"}, "To: email1@example.com,email3@example.com,email2@example.com"},
	// -t email1@example.com -t "email2@example.com, email3@example.com" -t email4@example.com
	{"multiple_flag_comma2", []string{"email1@example.com", "email2@example.com, email3@example.com", "email4@example.com"}, "To: email1@example.com,email2@example.com,email4@example.com,email3@example.com"},
}

func TestNewRcpts(t *testing.T) {
	for _, tc := range tcRcpts {
		t.Run(tc.name, func(t *testing.T) {

			r := newRcpts(tc.to)
			assert.Equal(t, tc.expTo, fmt.Sprintf("To: %s", r), "receipients should be equal")
		})
	}
}

func TestSendEmail(t *testing.T) {
	event := corev2.FixtureEvent("foo", "bar")
	//executed := time.Unix(event.Check.Executed, 0)
	//executedFormatted := executed.Format("2 Jan 2006 15:04:05")
	maps := make(map[string]string)
	maps["subject"] = "WebEx Monitoring Alert"
	maps["body"] = "<table width='960' cellspacing='0' cellpadding='0' border='0'><tbody><tr><td class='title_td'><span class='title_td_txt' lang='EN-US'>WebEx Monitoring Alert Summary</span> </td></tr></tbody></table><h3 class='table_title'>Time Range: [2020-11-26 11:11:36 - 2020-11-26 11:18:05]</h3><h3 class='table_title'>Duration:6 mins</h3><table cellspacing='0' cellpadding='0' border='0' style='BORDER-COLLAPSE: collapse'><thead><tr><th class='major_thead_td' nowrap=''><span class='major_thead_txt' lang='EN-US'>Test Target</span></th><th class='major_thead_td' nowrap=''><span class='major_thead_txt' lang='EN-US'>Alert Time</span></th><th class='major_thead_td' nowrap=''><span class='major_thead_txt' lang='EN-US'>Alert Count</span></th><th class='major_thead_td' nowrap=''><span class='major_thead_txt' lang='EN-US'>Alert Message</span></th><th class='major_thead_td' nowrap=''><span class='major_thead_txt' lang='EN-US'>Comment</span></th></tr></thead><tbody><tr><td style='height:20px'><span class='tbody_td_txt' lang='EN-US'>msj1mcs102.webex.com</span></td><td><span class='tbody_td_txt' lang='EN-US'>2020-11-26 11:14:36</span></td><td><span class='tbody_td_txt'>{MCT=1}</span></td><td><span class='tbody_td_txt' lang='EN-US'>{MCT=msj1mcs101.webex.com tested fail.}</span></td><td><span class='tbody_td_txt' lang='EN-US'>pool:msj1(mmp)</span></td></tr><tr><td style='height:20px'><span class='tbody_td_txt' lang='EN-US'>msj1mcs101.webex.com</span></td><td><span class='tbody_td_txt' lang='EN-US'>2020-11-26 11:11:36</span></td><td><span class='tbody_td_txt'>{MCT=1}</span></td><td><span class='tbody_td_txt' lang='EN-US'>{MCT=msj1mcs101.webex.com tested fail.}</span></td><td><span class='tbody_td_txt' lang='EN-US'>pool:msj1(mmp)</span></td></tr><tr><td class='table_sep_line' colspan='5'></td></tr></tbody></table>"
	event.Check.Annotations = maps
	config.SmtpHost = "mda.webex.com"
	config.Insecure = true
	config.FromEmail ="sensu@cisco.com"
	config.BodyTemplateFile = "/etc/sensu/email_template"
	config.ToEmail = []string{"jinjhe@cisco.com"}
	config.SubjectTemplate = "{{.Check.Annotations.subject}}"
	_ = checkArgs(event)
	_ = sendEmail(event)
}

func TestResolveBodyTemplate(t *testing.T) {
	event := corev2.FixtureEvent("foo", "bar")
	executed := time.Unix(event.Check.Executed, 0)
	executedFormatted := executed.Format("2 Jan 2006 15:04:05")
	template := "Entity: {{.Entity.Name}} Check: {{.Check.Name}} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}}"
	templout, err := resolveTemplate(template, event, "text/plain")
	assert.NoError(t, err)
	expected := fmt.Sprintf("Entity: foo Check: bar Executed: %s", executedFormatted)
	assert.Equal(t, templout, expected)
	template = "<html>Entity: {{.Entity.Name}} Check: {{.Check.Name}} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}}</html>"
	templout, err = resolveTemplate(template, event, "text/html")
	assert.NoError(t, err)
	expected = fmt.Sprintf("<html>Entity: foo Check: bar Executed: %s</html>", executedFormatted)
	assert.Equal(t, templout, expected)
}
