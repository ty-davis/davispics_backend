package main

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "net/smtp"
    "os"

    "github.com/joho/godotenv"
)


type Submittable interface {
    ToEmailBytes() []byte
    GetCaptcha() string
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    http.HandleFunc("/", serveForm)
    http.HandleFunc("/submit", handleBooking)
    http.HandleFunc("/question", handleQuestion)

    log.Println("Server starting on :8081")
    log.Fatal(http.ListenAndServe(":8081", nil))
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "https://davispics.com")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleBooking(w http.ResponseWriter, r* http.Request) {
    log.Println("WE ARE HERE 1")
    var booking Booking
    handleSubmit(w, r, &booking, "submit")
}

func handleQuestion(w http.ResponseWriter, r *http.Request) {
    log.Println("WE ARE HERE 2")
    var question Question
    handleSubmit(w, r, &question, "question")
}

func handleSubmit(w http.ResponseWriter, r *http.Request, target Submittable, action string) {
    log.Println("WE ARE HERE 3")
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


    if err := json.Unmarshal(bodyBytes, &target); err != nil {
        log.Printf("Invalid JSON: %v", err)
        http.Error(w, "Invalid JSON", 400)
        return
    }

    // Verify reCAPTCHA using Google Cloud reCAPTCHA Enterprise
    projectID := "davis-pictures-1749307333502"
    recaptchaKey := os.Getenv("RECAPTCHA_SITE_KEY")
    if err := CreateAssessment(projectID, recaptchaKey, target.GetCaptcha(), action); err != nil {
        log.Printf("reCAPTCHA verification failed: %v", err)
        http.Error(w, "reCAPTCHA verification failed", 400)
        return
    }

    // Send email
    if err := sendEmail(target.ToEmailBytes()); err != nil {
        log.Printf("Failed to send email: %v", err)
        http.Error(w, "Failed to send booking", 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}


func sendEmail(msg []byte) error {
    auth := smtp.PlainAuth("",
        os.Getenv("SMTP_EMAIL"),
        os.Getenv("SMTP_PASSWORD"),
        "smtp.gmail.com")
    to := []string{os.Getenv("SMTP_EMAIL")}

    return smtp.SendMail("smtp.gmail.com:587", auth, os.Getenv("SMTP_EMAIL"), to, msg)

}

func serveForm(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Booking API is running"))
}
