package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/smtp"
    "os"

    "github.com/joho/godotenv"
)

type Booking struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    FirstDateTime string `json:"first_datetime"`
    SecondDateTime string `json:"second_datetime"`
    ThirdDateTime string `json:"third_datetime"`
    Type     string `json:"type"`
    Comments string `json:"comments"`
    Captcha  string `json:"captcha"`
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    http.HandleFunc("/", serveForm)
    http.HandleFunc("/submit", handleSubmit)

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method == "OPTIONS" {
        return
    }

    if r.Method != "POST" {
        http.Error(w, "Method not allowed", 405)
        return
    }

    // Read the request body
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        log.Printf("Could not read request body: %v", err)
        http.Error(w, "Unable to read request", 400)
        return
    }
    defer r.Body.Close()

    log.Printf("Received raw body: %s", string(bodyBytes))

    var booking Booking
    if err := json.Unmarshal(bodyBytes, &booking); err != nil {
        log.Printf("Invalid JSON: %v", err)
        http.Error(w, "Invalid JSON", 400)
        return
    }
    log.Printf("Successfully decoded JSON: %+v", booking)

    // Verify reCAPTCHA using Google Cloud reCAPTCHA Enterprise
    projectID := "davis-pictures-1749307333502"
    recaptchaKey := os.Getenv("RECAPTCHA_SITE_KEY")
    if err := CreateAssessment(projectID, recaptchaKey, booking.Captcha, "submit"); err != nil {
        log.Printf("reCAPTCHA verification failed: %v", err)
        http.Error(w, "reCAPTCHA verification failed", 400)
        return
    }

    // Send email
    if err := sendEmail(booking); err != nil {
        log.Printf("Failed to send email: %v", err)
        http.Error(w, "Failed to send booking", 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func sendEmail(booking Booking) error {
    auth := smtp.PlainAuth("",
        os.Getenv("SMTP_EMAIL"),
        os.Getenv("SMTP_PASSWORD"),
        "smtp.gmail.com")

    to := []string{os.Getenv("SMTP_EMAIL")}
    msg := []byte(fmt.Sprintf(
        "Subject: New Booking Request\n\n"+
            "Name: %s\n"+
            "Email: %s\n"+
            "Phone: %s\n"+
            "First Date/Time: %s\n"+
            "Second Date/Time: %s\n"+
            "Third Date/Time: %s\n"+
            "Type: %s\n\n"+
            "Comments:\n%s\n",

        booking.Name, booking.Email, booking.Phone, booking.FirstDateTime, booking.SecondDateTime, booking.ThirdDateTime, booking.Type, booking.Comments))

    return smtp.SendMail("smtp.gmail.com:587", auth, os.Getenv("SMTP_EMAIL"), to, msg)
}

func serveForm(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Booking API is running"))
}
