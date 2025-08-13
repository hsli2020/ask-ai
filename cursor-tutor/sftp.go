package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	// read config.json
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	// connect to sftp server
	configSSH := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "sftp.example.com:22", configSSH)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// download files
	srcFile, err := client.Open("/remote/path/to/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File downloaded successfully!")
}

func SftpDownloadFile(config Config, remoteFilePath string, localFilePath string) error {
    // connect to sftp server
    configSSH := &ssh.ClientConfig{
        User: config.Username,
        Auth: []ssh.AuthMethod{
            ssh.Password(config.Password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    conn, err := ssh.Dial("tcp", "sftp.example.com:22", configSSH)
    if err != nil {
        return err
    }
    defer conn.Close()

    client, err := sftp.NewClient(conn)
    if err != nil {
        return err
    }
    defer client.Close()

    // download files
    srcFile, err := client.Open(remoteFilePath)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(localFilePath)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    _, err = srcFile.WriteTo(dstFile)
    if err != nil {
        return err
    }

    return nil
}

import (
	"bytes"
	"fmt"
	"net/smtp"
	"path/filepath"
)

type Email struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	Attachments []string
}

func SendEmail(email Email, smtpServer string, smtpPort int, smtpUsername string, smtpPassword string) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)

	// Set up the message.
	msg := "From: " + email.From + "\n" +
		"To: " + commaSeparatedString(email.To) + "\n" +
		"Cc: " + commaSeparatedString(email.Cc) + "\n" +
		"Bcc: " + commaSeparatedString(email.Bcc) + "\n" +
		"Subject: " + email.Subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: multipart/mixed; boundary=\"boundary\"\n\n" +
		"--boundary\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\n\n" +
		email.Body + "\n\n"

	for _, attachment := range email.Attachments {
		// Open the file.
		fileBytes, err := ioutil.ReadFile(attachment)
		if err != nil {
			return err
		}

		// Get the filename.
		filename := filepath.Base(attachment)

		// Add the attachment to the message.
		msg += "--boundary\n" +
			"Content-Type: application/octet-stream\n" +
			"Content-Disposition: attachment; filename=\"" + filename + "\"\n\n" +
			string(fileBytes) + "\n\n"
	}

	msg += "--boundary--"

	// Send the message.
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpServer, smtpPort), auth, email.From, append(email.To, email.Cc...), []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

func commaSeparatedString(strs []string) string {
	var buffer bytes.Buffer

	for i, str := range strs {
		if i > 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteString(str)
	}

	return buffer.String()
}

func ReceiveEmailsFromServer(email Email, server string, port int, username string, password string) ([]string, error) {
    // Set up authentication information.
    auth := smtp.PlainAuth("", username, password, server)

    // Connect to the server.
    client, err := pop3.Dial(fmt.Sprintf("%s:%d", server, port))
    if err != nil {
        return nil, err
    }
    defer client.Quit()

    // Log in to the server.
    err = client.Auth(username, password)
    if err != nil {
        return nil, err
    }

    // Get the number of messages.
    numMessages, err := client.Stat()
    if err != nil {
        return nil, err
    }

    // Get the messages.
    messages := make([]string, 0, numMessages)
    for i := 1; i <= numMessages; i++ {
        // Get the message.
        msg, err := client.Retr(i)
        if err != nil {
            return nil, err
        }

        // Convert the message to a string.
        msgStr := string(msg)

        // Add the message to the list of messages.
        messages = append(messages, msgStr)
    }

    return messages, nil
}
