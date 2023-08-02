package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"parser/types"
	"parser/utils"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/playwright-community/playwright-go"
)

var cookie playwright.BrowserContextAddCookiesOptionsCookies
var runner *playwright.Playwright
var browser playwright.Browser
var vk_api_token string = "vk1.a.brPshJonYQrHPmyYIYRBCZ3qAQ7mD5Go-xsUnzUKYqc1NHqPIupq1PjcA-wgb2_IAIHsrkm--2qZsVYU-Guwlc3YL1M5ZUE_pqo_1jgDFSRYvVXvn18PoQV4UBA6Dyj3yw96KN3xK-YzRNfLHT32br-Uox0Xu1m7_nXXNx_-jlcUyLREVABKk60J_VPvy75SZwzZM4ezzAon8WxbnZIwhA"
var pagesCount int
var period int

func main() {
	pagesCount = 4
	period = 60
	err := playwright.Install()
	utils.ErrorLog(err, "")
	runner, _ = playwright.Run()
	fls := false
	browser, _ = runner.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &fls,
	})
	cookie = createCookie()
	for {
		log.Println("iter")
		posts := parseMatchingPosts()
		db := getDB()
		var dbIds []uint32
		if len(db.Posts) != 0 {
			for _, dbPost := range db.Posts {
				dbIds = append(dbIds, dbPost.Id)
			}
		}

		for _, post := range posts {
			if utils.Contains(&dbIds, post.Id) {
				continue
			} else {
				msgStr := fmt.Sprintf("id: %d\ntitle: %s\nprice: %s\ndescription: %s\n",
					post.Id, post.Title, post.Price, post.Description)
				msgID := sendMessage(msgStr)
				addToDB(post.Id, msgID)
			}
		}

		dbToDeletePosts := []uint32{}
		messagesToDelete := []uint32{}

		if len(db.Posts) != 0 {
			for _, post := range db.Posts {

				if time.Now().Unix()-post.Time == 123 {
					print()
				} else if !checkPost(post.Id) {
					dbToDeletePosts = append(dbToDeletePosts, post.Id)
					messagesToDelete = append(messagesToDelete, post.Message_id)
				}
			}
			removePostsFromDB(dbToDeletePosts)
			deleteMessages(messagesToDelete)
		}

		time.Sleep(time.Minute)
	}

}

func addToDB(id uint32, message_id uint32) {
	db, err := sql.Open("sqlite3", "store.db")
	utils.ErrorLog(err, "")
	result, err := db.Exec("insert into posts (id, message_id, time) values ($1, $2, $3)", id, message_id, fmt.Sprint(time.Now().Unix()))
	utils.ErrorLog(err, "")
	fmt.Println(result.RowsAffected()) // количество добавленных строк
}
func removePostsFromDB(ids []uint32) {
	IDs := utils.UintSliceToString(ids)
	db, err := sql.Open("sqlite3", "store.db")
	utils.ErrorLog(err, "")
	r, err := db.Exec(fmt.Sprintf("delete from posts where id in (%s)", IDs))
	utils.ErrorLog(err, "")
	fmt.Println(r.RowsAffected()) // количество добавленных строк
}

func checkPost(id uint32) bool {
	context, _ := browser.NewContext()
	err := context.AddCookies(cookie)
	utils.ErrorLog(err, "")
	page, _ := context.NewPage()
	p := fmt.Sprintf("https://kwork.ru/projects/%d", id)
	_, err = page.Goto(p)
	utils.ErrorLog(err, "")
	_, err = page.WaitForSelector(".project-card")
	utils.ErrorLog(err, "")
	card, _ := page.QuerySelector(".project_card--informers_row")
	divs, _ := card.QuerySelectorAll(".mr15")
	offers_text, _ := divs[1].InnerText()
	offers_text = strings.Split(offers_text, "Предложений:")[1]
	offers_text = strings.ReplaceAll(offers_text, "\u00a0", "")
	offers, _ := strconv.Atoi(offers_text)
	fmt.Println(offers)

	if offers == 0 {
		return true
	} else {
		return false
	}

}

func getDB() types.DBPosts_T {
	var posts types.DBPosts_T
	db, err := sql.Open("sqlite3", "store.db")
	utils.ErrorLog(err, "")

	defer db.Close()
	rows, err := db.Query("SELECT * FROM posts")
	utils.ErrorLog(err, "")
	defer rows.Close()
	utils.ErrorLog(err, "")

	for rows.Next() {
		var post types.DBPost
		err := rows.Scan(&post.Id, &post.Message_id, &post.Time)
		utils.ErrorLog(err, "")
		posts.Add(post)
	}

	return posts
}

func createCookie() playwright.BrowserContextAddCookiesOptionsCookies {
	name, path, value, domain := "slrememberme",
		"/", "13813960_%242y%2410%24yuAfOB7e9n5.zLsANKsYke3W3nRgBMZ5UqlQNMhBpS89Xl16VerpO", "kwork.ru"

	cookie := playwright.BrowserContextAddCookiesOptionsCookies{
		Domain: &domain,
		Value:  &value,
		Name:   &name,
		Path:   &path,
	}
	return cookie
}

func deleteMessages(ids []uint32) {
	api_url := "https://api.vk.com/method/messages.delete"
	IDs := utils.UintSliceToString(ids)
	rand := rand.Uint32()
	_, err := http.Get(fmt.Sprintf("%s?access_token=%s&v=5.131&peer_id=251519327&message_ids=%s&delete_for_all=1&random_id=%d",
		api_url, vk_api_token, IDs, rand))
	utils.ErrorLog(err, "")
}

func sendMessage(message string) uint32 {
	api_url := "https://api.vk.com/method/messages.send"
	var id types.SentMsgID
	rand := rand.Uint32()
	msg := url.QueryEscape(message)
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s&v=5.131&user_id=251519327&message=%s&random_id=%d",
		api_url, vk_api_token, msg, rand))
	utils.ErrorLog(err, "")

	buf, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(buf, &id)
	utils.ErrorLog(err, "")
	fmt.Println(id.ID)
	return id.ID
}

func parseMatchingPosts() []types.Post {
	var posts []types.Post
	context, _ := browser.NewContext()
	err := context.AddCookies(cookie)
	utils.ErrorLog(err, "")

	for i := 1; i < pagesCount; i++ {
		page, _ := context.NewPage()
		p := fmt.Sprintf("https://kwork.ru/projects?page=%d", i)
		_, err := page.Goto(p)
		utils.ErrorLog(err, "")

		_, err = page.WaitForSelector("div.want-card")
		utils.ErrorLog(err, "")

		cards, _ := page.QuerySelectorAll("div.want-card")

		for _, card := range cards {
			infoBlock, _ := card.QuerySelector("div.query-item__info")
			text, _ := infoBlock.InnerText()
			offers_block := strings.Split(text, "Предложений: ")
			offers, _ := strconv.Atoi(offers_block[1])
			if offers != 0 {
				continue
			}

			titleBlock, _ := card.QuerySelector(".wants-card__header-title")
			title, _ := titleBlock.InnerText()

			id_text, _ := card.GetAttribute("data-id")
			id, _ := strconv.Atoi(id_text)

			descrBlock, _ := card.QuerySelector("div.wants-card__description-text")
			span, err := descrBlock.QuerySelector("span")
			utils.ErrorLog(err, "")
			if span != nil {
				span.Click()
			}
			descr, _ := descrBlock.InnerText()

			priceBlock, _ := card.QuerySelector("div.wants-card__header-price")
			spans, _ := priceBlock.QuerySelectorAll("span")
			price, _ := spans[1].InnerText()
			fmt.Printf("%d\n", id)
			posts = append(posts, types.Post{
				Id:          uint32(id),
				Title:       title,
				Description: descr,
				Price:       price,
			})
		}
		page.Close()
	}
	return posts

}
