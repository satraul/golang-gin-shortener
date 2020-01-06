package main

// only need mysql OR sqlite
// both are included here for reference
import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/speps/go-hashids"
)

var db *gorm.DB
var err error
var h *hashids.HashID

// Link model
type Link struct {
	ID   uint   `json:"id"`
	Link string `json:"link"`
	Slug string `json:"slug"`
}

func main() {
	// Initiate logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: false,
		},
	)

	// Initiate Hashids
	hd := hashids.NewData()
	hd.Salt = "this is my salt"
	h, _ = hashids.NewWithData(hd)
	// NOTE: See we’re using = to assign the global var
	// instead of := which would assign it only in this function
	db, err = gorm.Open("sqlite3", "./gorm.db")
	// db, _ = gorm.Open("mysql", "user:pass@tcp(127.0.0.1:3306)/database?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	db.AutoMigrate(&Link{})
	r := gin.New()
	r.Use(logger.SetLogger())
	r.Use(gin.Recovery())
	r.GET("/u/:slug", GetLink)
	r.POST("/u", CreateLink)
	r.Run(":8000")
}

// CreateLink creates link from body
func CreateLink(c *gin.Context) {
	var link Link
	c.BindJSON(&link)
	db.Create(&link)
	e, _ := h.Encode([]int{int(link.ID)})
	link.Slug = e
	db.Save(&link)
	link.Slug = GetAbsolute(c.Request) + "/" + link.Slug
	c.JSON(200, link)
}

// GetLink gets link from slug
func GetLink(c *gin.Context) {
	slug := c.Params.ByName("slug")
	d, err := h.DecodeWithError(slug)
	if err != nil {
		c.AbortWithStatus(404)
		return
	}
	id := d[0]
	var link Link
	if err := db.Where("id = ?", id).First(&link).Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
	} else {
		c.Redirect(301, link.Link)
	}
}

// GetAbsolute accepts an http.Request and returns it's absolute request URL
func GetAbsolute(r *http.Request) string {
	return fmt.Sprintf("%s%s%s", "http://", r.Host, r.URL.Path) //NOTE: http shouldn't be explicit
}
