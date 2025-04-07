package main

import (
	"middleware/example/internal/db"
	"middleware/example/internal/consumer"
)


func main() {

	db.InitDB()
	
	consumer.InitNats()
	
	consumer.RunMyConsumer()
		 
}

type MailMatter struct {
    Subject string `yaml:"subject"`
}

func GetStringFromEmbeddedTemplate(templatePath string, body interface{}) (content string, mailMatter MailMatter, err error) {
    temp, err := template.ParseFS(embeddedTemplates, templatePath)
    if err != nil {
        return
    }

    var tpl bytes.Buffer
    if err = temp.Execute(&tpl, body); err != nil {
        return
    }

    var mailContent []byte
    mailContent, err = frontmatter.Parse(strings.NewReader(tpl.String()), &mailMatter)
    if err == nil {
        content = string(mailContent)
    }

    return
}

func SendMail(to string, subject string, content string, token string) error {
    payload := map[string]interface{}{
        "to":      []string{to},
        "subject": subject,
        "content": content,
    }

    jsonBody, _ := json.Marshal(payload)
    req, err := http.NewRequest("POST", "https://mail.edu.forestier.re/api/mail", bytes.NewBuffer(jsonBody))
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("erreur envoi mail : %s", string(body))
    }

    return nil
}
