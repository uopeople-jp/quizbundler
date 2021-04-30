package quizbundler

import (
	"bytes"
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/k-yomo/moodle"
	"net/url"
	"time"
)

type Question struct {
	QuestionText     string
	Prompt           string
	Choices          []string
	SpecificFeedBack string
	GeneralFeedback  string
	RightAnswer      string
}

func BundleQuiz(ctx context.Context, loginParams moodle.LoginParams, courseID int) (map[string]*Question, error) {
	serviceURL, err := url.Parse("https://my.uopeople.edu")
	if err != nil {
		panic(err)
	}
	moodleClient, err := moodle.NewClientWithLogin(
		ctx,
		serviceURL,
		&moodle.LoginParams{
			Username: loginParams.Username,
			Password: loginParams.Password,
		},
	)
	if err != nil {
		return nil, err
	}

	quizzes, err := moodleClient.QuizAPI.GetQuizzesByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// we use map to remove duplicates,
	questionMap := make(map[string]*Question)

	for _, q := range quizzes {
		attempts, err := moodleClient.QuizAPI.GetUserAttempts(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		for _, attempt := range attempts {
			time.Sleep(1 * time.Second)
			_, questions, err := moodleClient.QuizAPI.GetAttemptReview(ctx, attempt.ID)
			if err != nil {
				return nil, err
			}
			for _, question := range questions {
				q, err := extractQuestionFromHTML(question.HtmlRaw)
				if err != nil {
					return nil, err
				}
				questionMap[q.QuestionText] = q
			}
		}
	}

	return questionMap, nil
}

func extractQuestionFromHTML(questionHTML string) (*Question, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(questionHTML)))
	if err != nil {
		return nil, err
	}

	var choices []string
	doc.Find("label.ml-1").Each(func(i int, s *goquery.Selection) {
		choices = append(choices, s.Text())
	})

	return &Question{
		QuestionText:     doc.Find("div.qtext").Text(),
		Prompt:           doc.Find("div.prompt").Text(),
		Choices:          choices,
		SpecificFeedBack: doc.Find("div.specificfeedback").Text(),
		GeneralFeedback:  doc.Find("div.generalfeedback").Text(),
		RightAnswer:      doc.Find("div.rightanswer").Text(),
	}, nil
}
