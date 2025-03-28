package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message
type Message struct {
	ID        string
	Content   string
	Timestamp time.Time
	UserID    string
}

// UserSession tracks anonymous user activity
type UserSession struct {
	ID             string
	MessagesPosted int
	LastActive     time.Time
}

// Global application state
var (
	messages     = []Message{}
	userSessions = make(map[string]*UserSession)
	mutex        = sync.RWMutex{}
)

// Templates
var templates = template.Must(template.ParseFiles("templates/index.html"))

// Main page handler
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get or create user ID
	userID := ensureUserID(w, r)
	messagesRemaining := getRemainingMessages(userID)
	needsRegistration := messagesRemaining <= 0

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"MessagesRemaining": messagesRemaining,
		"NeedsRegistration": needsRegistration,
	})
}

// Get messages handler
func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, msg := range messages {
		timeStr := msg.Timestamp.Format("15:04")
		fmt.Fprintf(w, `
		<div class="message-item bg-gray-800/80 p-3 rounded-lg mb-2 backdrop-blur-sm">
			<div class="flex justify-between items-start">
				<p class="text-gray-200 break-words">%s</p>
				<span class="text-xs text-gray-500 ml-2">%s</span>
			</div>
		</div>
		`, msg.Content, timeStr)
	}
}

// Post message handler
func postMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := ensureUserID(w, r)
	remaining := getRemainingMessages(userID)

	if remaining <= 0 {
		w.Header().Set("HX-Trigger", `{"showRegistration": true}`)
		fmt.Fprint(w, `
		<div class="flex justify-center items-center h-32">
			<div class="text-red-400 text-center">
				<p>You've reached your limit of 5 anonymous messages</p>
				<button class="mt-2 text-shadow-400 underline" hx-get="/register-prompt" hx-target="#modal-container">
					Create an account to continue
				</button>
			</div>
		</div>
		`)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// Create new message
	newMsg := Message{
		ID:        uuid.New().String(),
		Content:   message,
		Timestamp: time.Now(),
		UserID:    userID,
	}

	// Add to messages
	mutex.Lock()
	messages = append(messages, newMsg)
	
	// Update user session
	if session, exists := userSessions[userID]; exists {
		session.MessagesPosted++
		session.LastActive = time.Now()
	}
	mutex.Unlock()

	// Return all messages and update counter
	getMessagesHandler(w, r)
	remaining--
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"updateCounter": "%d"}`, remaining))
}

// Register prompt handler
func registerPromptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
	<div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50">
		<div class="bg-gray-800 p-6 rounded-xl border border-gray-700 max-w-md w-full shadow-xl">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-bold text-white">Create an Account</h2>
				<button hx-get="/close-modal" hx-target="#modal-container" class="text-gray-400 hover:text-white">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
					</svg>
				</button>
			</div>
			
			<p class="text-gray-300 mb-4">Join ShadowSpeak to unlock unlimited messaging and more features.</p>
			
			<form class="space-y-4">
				<div>
					<label for="username" class="block text-gray-300 mb-1 text-sm">Username</label>
					<input type="text" id="username" class="w-full bg-gray-700 border border-gray-600 rounded-md px-3 py-2 text-gray-200 focus:outline-none focus:ring-2 focus:ring-shadow-500">
				</div>
				<div>
					<label for="email" class="block text-gray-300 mb-1 text-sm">Email</label>
					<input type="email" id="email" class="w-full bg-gray-700 border border-gray-600 rounded-md px-3 py-2 text-gray-200 focus:outline-none focus:ring-2 focus:ring-shadow-500">
				</div>
				<div>
					<label for="password" class="block text-gray-300 mb-1 text-sm">Password</label>
					<input type="password" id="password" class="w-full bg-gray-700 border border-gray-600 rounded-md px-3 py-2 text-gray-200 focus:outline-none focus:ring-2 focus:ring-shadow-500">
				</div>
				
				<div class="pt-2">
					<button type="button" class="w-full px-4 py-2 bg-shadow-600 hover:bg-shadow-700 text-white rounded-md transition-colors duration-200 shadow-lg flex justify-center items-center">
						<span>Create Account</span>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 ml-1" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 1.414L10.586 9H7a1 1 0 100 2h3.586l-1.293 1.293a1 1 0 101.414 1.414l3-3a1 1 0 000-1.414z" clip-rule="evenodd" />
						</svg>
					</button>
				</div>
			</form>
		</div>
	</div>
	`)
}

// Close modal handler
func closeModalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "")
}

// Helper function to get/create user ID
func ensureUserID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("user_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	
	// Create new user ID
	userID := uuid.New().String()
	
	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30, // 30 days
	})
	
	// Create user session
	mutex.Lock()
	userSessions[userID] = &UserSession{
		ID:             userID,
		MessagesPosted: 0,
		LastActive:     time.Now(),
	}
	mutex.Unlock()
	
	return userID
}

// Get remaining messages for user
func getRemainingMessages(userID string) int {
	mutex.RLock()
	defer mutex.RUnlock()
	
	session, exists := userSessions[userID]
	if !exists {
		return 5
	}
	
	remaining := 5 - session.MessagesPosted
	if remaining < 0 {
		return 0
	}
	return remaining
}

func main() {
	// Register routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/get-messages", getMessagesHandler)
	http.HandleFunc("/post-message", postMessageHandler)
	http.HandleFunc("/register-prompt", registerPromptHandler)
	http.HandleFunc("/close-modal", closeModalHandler)
	
	// Initialize sample messages
	messages = append(messages, Message{
		ID:        uuid.New().String(),
		Content:   "Welcome to ShadowSpeak! Share your thoughts anonymously.",
		Timestamp: time.Now(),
		UserID:    "system",
	})
	
	// Start server
	port := "8080"
	fmt.Printf("Starting ShadowSpeak on port %s...\n", port)
	fmt.Printf("Visit http://localhost:%s in your browser.\n", port)
	
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
