package main

import (
	"github.com/astianmuchui/go-auth/auth"
	"github.com/astianmuchui/go-auth/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/django/v3"
	"log"
)

var store = session.New()

func main() {

	engine := django.New("./templates", ".django")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	/* Home route */
	app.Get("/", func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		errorMsg := sess.Get("signup_error")
		sess.Delete("signup_error")
		sess.Save()

		return c.Render("index", fiber.Map{
			"signup_error": errorMsg,
		})
	})

	app.Post("/register", func(context *fiber.Ctx) error {
		payload := new(models.User)

		if err := context.BodyParser(payload); err != nil {
			return err
		}
		sess, _ := store.Get(context)

		// Create the user if already does not exist
		if models.UserExists(payload) == false {
			// Add user to database
			result := models.CreateUser(payload)

			if result.Error != nil {
				return context.SendStatus(fiber.StatusCreated)
			} else {
				sess.Set("signup_error", "Unable to sign up")
				sess.Save()
			}
		} else {
			sess.Set("signup_error", "User already exists")
			log.Println("User found")
			sess.Save()
			return context.Redirect("/")
		}

		log.Println("Username:", payload.Username)
		log.Println("Email:", payload.Email)
		log.Println("Password:", payload.Password)

		return context.SendString("Username: " + payload.Username + " Email: " + payload.Email + " Password: " + payload.Password)
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		errorMsg := sess.Get("login_error")
		sess.Delete("login_error")
		sess.Save()

		return c.Render("login", fiber.Map{
			"login_error": errorMsg,
		})
	})

	app.Post("/signin", func(c *fiber.Ctx) error {
		payload := new(models.User)

		if err := c.BodyParser(payload); err != nil {
			return err
		}

		var u models.User
		u = models.GetUserDataByUsername(payload.Username)
		log.Println(u)
		log.Println(payload)

		userVerified := auth.Login(payload)
		if userVerified {
			sess, _ := store.Get(c)
			sess.Set("login_error", "Invalid username or password")
			sess.Set("user_email", u.Email)
			sess.Set("logged_in", true)

			sess.Save()

			return c.Redirect("/dashboard")
		}

		return c.Redirect("/login")
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {

		sess, _ := store.Get(c)

		logged_in := sess.Get("logged_in")
		log.Println(logged_in)
		if logged_in == true {
			userEmail := sess.Get("user_email").(string)
			log.Println(userEmail)
			user_data := models.GetUserDataByEmail(userEmail)
			return c.Render("dashboard", fiber.Map{
				"username": user_data.Username,
				"email":    user_data.Email,
			})
		} else {
			return c.Redirect("/login")
		}
	})

	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Destroy()
		return c.Redirect("/login")
	})

	app.Get("/update", func(c *fiber.Ctx) error {

		sess, _ := store.Get(c)
		logged_in := sess.Get("logged_in")

		if logged_in == true {
			userEmail := sess.Get("user_email").(string)
			log.Println(userEmail)
			user_data := models.GetUserDataByEmail(userEmail)

			return c.Render("update", fiber.Map{
				"username": user_data.Username,
				"email":    user_data.Email,
			})
		} else {
			return c.Redirect("/login")
		}
	})

	// app.Post("/update-profile", func(c *fiber.Ctx) error {
	// 	payload := new(models.User)

	// 	if err := c.BodyParser(payload); err != nil {
	// 		return err
	// 	}

		
	// })
	app.Listen(":8081")
}
