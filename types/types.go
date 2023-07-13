package types

import "sync"

type Post struct {
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
	Time       uint64
}

type DBPosts_T struct {
	Posts []DBPost
}

func (self *DBPosts_T) Add(post DBPost) {
	self.Posts = append(self.Posts, DBPost{
		Id:         post.Id,
		Message_id: post.Message_id,
		Time:       post.Time,
	})
}
