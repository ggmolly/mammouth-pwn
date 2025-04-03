package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/tiktoken-go/tokenizer"
)

var (
	defaultTextModel           = "anthropic-claude-3-7-sonnet-latest"
	defaultPricePerTokenInput  = float64(3e-06)
	defaultPricePerTokenOutput = float64(1.5e-05)

	minSleep = 30000 // 3 mins
	maxSleep = 90000 // 9 mins

	errMinSleep = 1800000 // 30 mins
	errMaxSleep = 5400000 // 90 mins
)

type MammouthConversationCreateDto struct {
	Message       string        `json:"message"`
	Model         string        `json:"model"`
	DefaultModels DefaultModels `json:"defaultModels"`
	AssistantID   interface{}   `json:"assistantId"`
	Attachments   []interface{} `json:"attachments"`
}

type DefaultModels struct {
	Text      string `json:"text"`
	Image     string `json:"image"`
	WebSearch string `json:"webSearch"`
}

type MammouthConversation struct {
	ID          int64       `json:"id"`
	AssistantID interface{} `json:"assistantId"`
	User        int64       `json:"user"`
	Title       string      `json:"title"`
	Model       string      `json:"model"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	PublicID    interface{} `json:"publicId"`
}

func (c *MammouthConversation) SendMessage(client *MammouthClient, message string) (string, error) {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	formField, _ := writer.CreateFormField("model")
	formField.Write([]byte(c.Model))
	formField, _ = writer.CreateFormField("preprompt")
	formField.Write([]byte(""))
	formField, _ = writer.CreateFormField("messages")
	formField.Write([]byte(fmt.Sprintf(`{"content":"%s","imagesData":[],"documentsData":[]}`, message)))
	writer.Close()
	req, _ := http.NewRequest("POST", "https://mammouth.ai/api/models/llms", form)
	setHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func (c *MammouthConversation) Delete(client *MammouthClient) error {
	data := strings.NewReader(fmt.Sprintf(`{"id":%d}`, c.ID))
	req, _ := http.NewRequest("POST", "https://mammouth.ai/api/chat/delete", data)
	setHeaders(req)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err
	}

	return nil
}

func setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://mammouth.ai/app/a/default")
	req.Header.Set("Origin", "https://mammouth.ai")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Priority", "u=0")
	req.Header.Set("TE", "trailers")
}

type MammouthClient struct {
	httpClient *http.Client
}

func (c *MammouthClient) CreateConversation(model, message string) (MammouthConversation, error) {
	dto := MammouthConversationCreateDto{
		Message:       message,
		Model:         model,
		DefaultModels: DefaultModels{Text: defaultTextModel, Image: "replicate-recraftai-recraft-v3", WebSearch: "openperplex-v1"},
		AssistantID:   nil,
		Attachments:   []interface{}{},
	}

	body, _ := json.Marshal(dto)

	req, _ := http.NewRequest("POST", "https://mammouth.ai/api/chat/create", bytes.NewBuffer(body))
	setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MammouthConversation{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return MammouthConversation{}, err
	}

	var conversation MammouthConversation
	err = json.NewDecoder(resp.Body).Decode(&conversation)
	if err != nil {
		return MammouthConversation{}, err
	}

	return conversation, nil
}

func NewMammouthClient(token, gcpToken, language string) *MammouthClient {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	cookies := []*http.Cookie{
		{
			Name:     "auth_session",
			Value:    token,
			Domain:   "mammouth.ai",
			Path:     "/",
			Secure:   true,
			HttpOnly: false,
		},
		{
			Name:     "i18n_redirected",
			Value:    language,
			Domain:   "mammouth.ai",
			Path:     "/",
			Secure:   false,
			HttpOnly: false,
		},
		{
			Name:     "gcp_token",
			Value:    gcpToken,
			Domain:   "mammouth.ai",
			Path:     "/",
			Secure:   false,
			HttpOnly: false,
		},
	}

	url, _ := url.Parse("https://mammouth.ai/")

	c.Jar, _ = cookiejar.New(nil)
	c.Jar.SetCookies(url, cookies)

	return &MammouthClient{
		httpClient: c,
	}
}

func tokenCost(input string, tokenizer tokenizer.Codec, isInput bool) (int, float64) {
	ids, _, _ := tokenizer.Encode(input)
	if isInput {
		return len(ids), float64(len(ids)) * defaultPricePerTokenInput
	} else {
		return len(ids), float64(len(ids)) * defaultPricePerTokenOutput
	}
}

func verboseSleep(message string, t time.Duration) {
	log.Println("[#] sleeping for", t.String(), ":", message)
	time.Sleep(t)
}

func main() {
	enc, _ := tokenizer.Get(tokenizer.O200kBase)
	client := NewMammouthClient(
		"7ygl6hom3wkcnvwbw5zs2wq6b6xlbrbwoykwoemy",
		"39228_1746115549092_4593d53f5399202e40e489d6dff97c7433a4402d0abd9976e3264e4f6670b76c",
		"en",
	)
	sessCostInput := float64(0)
	sessCostOutput := float64(0)

	for {
		conversation, err := client.CreateConversation(
			"anthropic-claude-3-7-sonnet-latest",
			"Hello, I'd you to tell me a long...",
		)
		if err == nil {
			message := "Hello, I'd like you to tell me a long story about a cat."
			response, msgErr := conversation.SendMessage(client, message)
			{
				tokenCount, cost := tokenCost(message, enc, true)
				log.Printf("[#] input tokens: %d, cost: %.4f\n", tokenCount, cost)
				sessCostInput += cost
			}
			{
				tokenCount, cost := tokenCost(response, enc, true)
				log.Printf("[#] output tokens: %d, cost: %.4f\n", tokenCount, cost)
				sessCostOutput += cost
			}
			conversation.Delete(client)
			log.Printf("Total session I/O cost: %.4f€\n", sessCostInput+sessCostOutput)
			log.Println("--------------------------")

			if msgErr != nil {
				log.Println("[-] Error sending message:", msgErr)
				verboseSleep("waiting for next cycle (error)", time.Duration(rand.Intn(errMaxSleep-errMinSleep)+errMinSleep)*time.Millisecond)
				continue
			} else {
				verboseSleep("waiting for next cycle", time.Duration(rand.Intn(maxSleep-minSleep)+minSleep)*time.Millisecond)
			}
		} else {
			log.Println("[-] Error creating conversation:", err)
			verboseSleep("waiting for next cycle (error)", time.Duration(rand.Intn(errMaxSleep-errMinSleep)+errMinSleep)*time.Millisecond)
		}
	}
}
