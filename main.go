package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"parser/types"
	"parser/utils"
	"strconv"
	"strings"
	"sync"

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
	pagesCount = 5
	period = 60
	runner, _ = playwright.Run()
	fls := false
	browser, _ = runner.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &fls,
	})
	log.Println("started")
	cookie = createCookie()
	x := getDB()
	for _, x := range x.Posts {
		fmt.Println(fmt.Sprintf("%d, %d, %d\n", x.Id, x.Message_id, x.Time))
	}

	/*
		for {

			time.Sleep(time.Duration(period) * time.Second)
		}
	*/
}

func getDB() types.DBPosts_T {
	var posts types.DBPosts_T
	db, err := sql.Open("sqlite3", "store.db")
	utils.ErrorLog(err, "")

	defer db.Close()
	rows, err := db.Query("SELECT * FROM posts")
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

func getPost(id types.PostID_T) types.Post {
	context, err := browser.NewContext()
	utils.ErrorLog(err, "")

	page, _ := context.NewPage()
	_, err = page.Goto(
		fmt.Sprintf("https://kwork.ru/projects/%d", id))
	utils.ErrorLog(err, "")

	log.Println("success")
	card, err := page.QuerySelector(".project-card")
	utils.ErrorLog(err, "")

	title := func() string {
		titleBlock, err := card.QuerySelector(".wants-card__header-title")
		utils.ErrorLog(err, "")

		title, err := titleBlock.InnerText()
		utils.ErrorLog(err, "")

		return title
	}()

	descr := func() string {
		block, err := card.QuerySelector(".wants-card__description-text")
		utils.ErrorLog(err, "")

		text, err := block.InnerText()
		utils.ErrorLog(err, "")

		return text
	}()

	price := func() string {
		block, err := card.QuerySelector(".wants-card__header-price")
		utils.ErrorLog(err, "")

		block2, err := block.QuerySelector(".d-inline")
		utils.ErrorLog(err, "")

		text, err := block2.InnerText()
		utils.ErrorLog(err, "")

		return text
	}()

	post := types.Post{
		Title:       title,
		Description: descr,
		Price:       price,
	}

	return post
}

func getPosts(postIds []types.PostID_T) types.PostsT {
	var posts types.PostsT
	var wg sync.WaitGroup
	wg.Add(len(postIds))
	for _, id := range postIds {
		go func(_id types.PostID_T) {
			defer wg.Done()
			post := getPost(_id)
			posts.Mutex.Lock()
			posts.Posts = append(posts.Posts, post)
			posts.Mutex.Unlock()
		}(id)
	}
	wg.Wait()
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

func sendMessage(message string) {
	api_url := "https://api.vk.com/method/messages.send"
	rand := rand.Uint32()

	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s&v=5.131&user_id=251519327&message=%s&random_id=%d",
		api_url, vk_api_token, message, rand))
	utils.ErrorLog(err, "")

	buf, _ := io.ReadAll(resp.Body)
	r := string(buf)
	log.Println(r)
}

func parseMatchingPosts() []types.PostID_T {
	var ids []types.PostID_T
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
			block, _ := card.QuerySelector("div.query-item__info")
			text, _ := block.InnerText()
			offers_block := strings.Split(text, "Предложений: ")
			offers, _ := strconv.Atoi(offers_block[1])
			if offers != 0 {
				continue
			}
			id_text, _ := card.GetAttribute("data-id")
			id, _ := strconv.Atoi(id_text)
			ids = append(ids, uint32(id))

		}
	}
	return ids

}
