package post

import (
	"sync"
	"time"

	"reddit/pkg/comment"
	"reddit/pkg/idgenerator"
	"reddit/pkg/user"
	"reddit/pkg/vote"
)

type PostDBRepository interface {
	IncreasePostViewsDB(post *Post, postID string) error
	GetPostByCategoryDB(postOfCurrentCategory []*Post, category string) ([]*Post, error)
	AddPostDB(post *Post) error
	GetAllPostsDB(allPosts []*Post) ([]*Post, error)
	AddCommentDB(post *Post, postID string) error
	DeleteCommentDB(postWithCommentToDelete *Post, postID string) error
	GetPostByIDDB(postID string) (*Post, error)
	SetPostDB(postToSet *Post, postID string) error
	GetPostByUsernameDB(userName string) ([]*Post, error)
	DeletePostDB(postID string) (bool, error)
}

type PostBusinessLogic struct {
	mu          *sync.RWMutex
	PostDBRepo  PostDBRepository
	generatorID idgenerator.IDGenerator
}

func NewPostBusinessLogic(repo PostDBRepository, idGenerator idgenerator.IDGenerator) *PostBusinessLogic {
	return &PostBusinessLogic{
		PostDBRepo:  repo,
		mu:          &sync.RWMutex{},
		generatorID: idGenerator,
	}
}

func getTimeOfCreation() string {
	timeOfCreation := time.Now()
	return timeOfCreation.Format("2006-01-02T15:04:05.999Z")

}

func (p *PostBusinessLogic) GetAll() ([]*Post, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	allPosts := make([]*Post, 0)
	allPosts, err := p.PostDBRepo.GetAllPostsDB(allPosts)
	if err != nil {
		return nil, err
	}
	return allPosts, nil
}

func (p *PostBusinessLogic) AddPost(post *Post, author *user.User) (*Post, error) {
	post.Author = author
	post.Votes = make([]*vote.Vote, 0, 1)
	post.Votes = append(post.Votes, &vote.Vote{
		Value:  1,
		UserID: author.ID,
	})
	if post.Type == "text" {
		post.URL = ""
	} else {
		post.Text = ""
	}
	post.Views = 0
	post.Comments = make([]*comment.Comment, 0)
	post.Created = getTimeOfCreation()
	post.UpvotePercentage = 100
	post.Score = 1
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.PostDBRepo.AddPostDB(post)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostBusinessLogic) GetPostByCategory(category string) ([]*Post, error) {
	postOfCurrentCategory := make([]*Post, 0)
	p.mu.RLock()
	defer p.mu.RUnlock()
	postOfCurrentCategory, err := p.PostDBRepo.GetPostByCategoryDB(postOfCurrentCategory, category)
	if err != nil {
		return nil, err
	}
	return postOfCurrentCategory, nil
}

func (p *PostBusinessLogic) GetPostByID(id string) (*Post, error) {
	post, err := p.findPostByID(id)
	if err != nil {
		return nil, err
	}
	post.Views++
	p.mu.Lock()
	defer p.mu.Unlock()
	err = p.PostDBRepo.IncreasePostViewsDB(post, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostBusinessLogic) AddComment(commentBody string, author *user.User, postID string) (*Post, error) {
	post, err := p.findPostByID(postID)
	if err != nil {
		return nil, err
	}
	newComment := &comment.Comment{
		Created: getTimeOfCreation(),
		Author:  author,
		Body:    commentBody,
		ID:      p.generatorID.GenerateID(16),
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	post.Comments = append(post.Comments, newComment)
	err = p.PostDBRepo.AddCommentDB(post, postID)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostBusinessLogic) DeleteComment(userID, postID, commentID string) (*Post, error) {
	postWithCommentToDelete, err := p.findPostByID(postID)
	if err != nil {
		return nil, ErrNoPost
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, currentComment := range postWithCommentToDelete.Comments {
		if currentComment.ID == commentID {
			if currentComment.Author.ID != userID {
				return nil, ErrNoAccess
			}
			postWithCommentToDelete.Comments = append(postWithCommentToDelete.Comments[:i], postWithCommentToDelete.Comments[i+1:]...)
			err = p.PostDBRepo.DeleteCommentDB(postWithCommentToDelete, postID)
			if err != nil {
				return nil, err
			}
			return postWithCommentToDelete, nil
		}
	}
	return nil, ErrNoComment

}

func (p *PostBusinessLogic) UpVote(postID string, userID string) (*Post, error) {
	postToUpvote, err := p.findPostByID(postID)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, currentVote := range postToUpvote.Votes {
		if currentVote.UserID == userID {
			if currentVote.Value == -1 {
				currentVote.Value = 1
				postToUpvote.Score += 2
				postToUpvote.UpvotePercentage = countUpVotePercentage(postToUpvote)
				err = p.PostDBRepo.SetPostDB(postToUpvote, postID)
				if err != nil {
					return nil, err
				}
			}
			return postToUpvote, nil
		}
	}
	postToUpvote.Votes = append(postToUpvote.Votes, vote.NewVote(1, userID))
	postToUpvote.Score++
	postToUpvote.UpvotePercentage = countUpVotePercentage(postToUpvote)
	err = p.PostDBRepo.SetPostDB(postToUpvote, postID)
	if err != nil {
		return nil, err
	}
	return postToUpvote, nil

}

func (p *PostBusinessLogic) DownVote(postID string, userID string) (*Post, error) {
	postToDownvote, err := p.findPostByID(postID)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, currentVote := range postToDownvote.Votes {
		if currentVote.UserID == userID {
			if currentVote.Value == 1 {
				postToDownvote.Score -= 2
				currentVote.Value = -1
				postToDownvote.UpvotePercentage = countUpVotePercentage(postToDownvote)
				err = p.PostDBRepo.SetPostDB(postToDownvote, postID)

				if err != nil {
					return nil, err
				}
			}
			return postToDownvote, nil
		}
	}
	postToDownvote.Votes = append(postToDownvote.Votes, vote.NewVote(-1, userID))
	postToDownvote.Score--
	postToDownvote.UpvotePercentage = countUpVotePercentage(postToDownvote)
	err = p.PostDBRepo.SetPostDB(postToDownvote, postID)
	if err != nil {
		return nil, err
	}
	return postToDownvote, nil

}

func (p *PostBusinessLogic) UnVote(postID string, userID string) (*Post, error) {
	postToUnvote, err := p.findPostByID(postID)
	if err != nil {
		return nil, ErrNoPost
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, currentVote := range postToUnvote.Votes {
		if currentVote.UserID == userID {
			postToUnvote.Votes = append(postToUnvote.Votes[:i], postToUnvote.Votes[i+1:]...)
			if len(postToUnvote.Votes) == 0 {
				postToUnvote.Score = 0
				postToUnvote.UpvotePercentage = 0
			} else {
				if currentVote.Value == 1 {
					postToUnvote.Score--
				} else {
					postToUnvote.Score++
				}
				postToUnvote.UpvotePercentage = countUpVotePercentage(postToUnvote)
			}
			err = p.PostDBRepo.SetPostDB(postToUnvote, postID)
			if err != nil {
				return nil, err
			}
			return postToUnvote, nil
		}
	}
	return postToUnvote, nil

}

func (p *PostBusinessLogic) DeletePost(userID, postID string) (bool, error) {
	postToDelete, err := p.findPostByID(postID)
	if err != nil {
		return false, ErrNoPost
	}
	if postToDelete.Author.ID != userID {
		return false, ErrNoAccess
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.PostDBRepo.DeletePostDB(postID)
}

func (p *PostBusinessLogic) GetPostsByUserID(userName string) ([]*Post, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	userPosts, err := p.PostDBRepo.GetPostByUsernameDB(userName)
	if err != nil {
		return nil, err
	}
	return userPosts, nil
}

func countUpVotePercentage(postToCount *Post) int {
	var numOfUpVotes int
	for _, currentVote := range postToCount.Votes {
		if currentVote.Value == 1 {
			numOfUpVotes++
		}
	}
	return 100 * numOfUpVotes / len(postToCount.Votes)
}

func (p *PostBusinessLogic) findPostByID(id string) (*Post, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	post, err := p.PostDBRepo.GetPostByIDDB(id)
	if err != nil {
		return nil, err
	}
	return post, nil
}
