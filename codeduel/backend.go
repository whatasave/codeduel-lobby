package codeduel

import (
	"fmt"

	"github.com/xedom/codeduel-lobby/codeduel/utils"
)

type Backend struct {
	apiBaseUrl string
	apiToken   string
}

func NewBackend(apiBaseUrl string, apiToken string) Backend {
	return Backend{
		apiBaseUrl: apiBaseUrl,
		apiToken:   apiToken,
	}
}

func (backend *Backend) get(path string, response interface{}) error {
	return utils.HttpGet(backend.apiBaseUrl+path, map[string]string{
		"x-token": backend.apiToken,
	}, response)
}

func (backend *Backend) post(path string, body any) (any, error) {
	var jsonResponse any
	if backend == nil {
		return nil, fmt.Errorf("backend is nil")
	}
	headers := map[string]string{
		"x-token": backend.apiToken,
	}
	err := utils.HttpPost(backend.apiBaseUrl+path, headers, body, &jsonResponse)
	if err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

func (backend *Backend) patch(path string, body any) (any, error) {
	var jsonResponse any
	headers := map[string]string{
		"x-token": backend.apiToken,
	}
	err := utils.HttpPatch(backend.apiBaseUrl+path, headers, body, &jsonResponse)
	if err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

func (backend *Backend) CreateLobby(lobby *Lobby) error {
	_, err := backend.post("/v1/lobby", map[string]any{
		"lobby_id":     lobby.Id,
		"owner_id":     lobby.Owner.Id,
		"users_id":     keys(lobby.Users),
		"challenge_id": lobby.State.(GameLobbyState).Challenge.Id,
		"settings":     lobby.Settings,
	})
	return err
}

func (backend *Backend) RegisterSubmission(lobby *Lobby, user User, runResult *RunResult) error {
	_, err := backend.patch("/v1/lobby/"+lobby.Id+"/submission", map[string]any{
		"user_id":      user.Id,
		"code":         runResult.Code,
		"language":     runResult.Language,
		"date":         runResult.Date,
		"tests_passed": runResult.PassedTests,
	})
	return err
}

func (backend *Backend) EndLobby(lobby *Lobby) error {
	_, err := backend.patch("/v1/lobby/"+lobby.Id+"/endgame", map[string]any{})
	return err
}

func (backend *Backend) GetChallenge(challengeId string) (*Challenge, error) {
	challenge := &Challenge{}
	err := backend.get("/v1/challenge/"+challengeId+"/full", challenge)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

func (backend *Backend) GetRandomChallenge() (*Challenge, error) {
	challenge := &Challenge{}
	err := backend.get("/v1/challenge/random", challenge)

	return challenge, err
}

func keys[K comparable, V any](dict map[K]V) []K {
	keys := make([]K, len(dict))
	i := 0
	for k := range dict {
		keys[i] = k
		i++
	}
	return keys
}
