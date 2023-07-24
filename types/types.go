package types

import "sync"

type Post struct {
	Id          uint32
	Title       string
	Price       string
	Description string
}

type PostID_T = uint32

type PostsT struct {
	*sync.Mutex
	Posts []Post
}

type DBPost struct {
	Id         PostID_T
	Message_id uint32
	Time       int64
}

type DBPosts_T struct {
	Posts []DBPost
}

func (s *DBPosts_T) Add(post DBPost) {
	s.Posts = append(s.Posts, DBPost{
		Id:         post.Id,
		Message_id: post.Message_id,
		Time:       post.Time,
	})
}

type SentMsgID struct {
	ID uint32 `json:"response"`
}
