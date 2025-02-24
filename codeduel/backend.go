package codeduel

import (
	"fmt"
	"log"
	"strings"

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
	log.Printf("Creating lobby %v", lobby)
	_, err := backend.post("/v1/game", map[string]any{
		"uniqueId":         lobby.Id,
		"ownerId":          lobby.Owner.Id,
		"users":            keys(lobby.Users),
		"challengeId":      lobby.State.(GameLobbyState).Challenge.Id,
		"modeId":           1,
		"ended":            false,
		"maxPlayers":       lobby.Settings.MaxPlayers,
		"allowedLanguages": strings.Join(lobby.Settings.AllowedLanguages, ","),
		"gameDuration":     lobby.Settings.GameDuration,
	})
	return err
}

func (backend *Backend) RegisterSubmission(lobby *Lobby, user User, runResult *RunResult) error {
	_, err := backend.patch("/v1/game/"+lobby.Id+"/submit", map[string]any{
		"userId":      user.Id,
		"gameId":      lobby.Id,
		"code":        runResult.Code,
		"language":    runResult.Language,
		"testsPassed": runResult.PassedTests,
		"submittedAt": runResult.Date.String(),
	})
	return err
}

func (backend *Backend) EndLobby(lobby *Lobby) error {
	_, err := backend.patch("/v1/game/"+lobby.Id+"/endgame", map[string]any{})
	return err
}

func (backend *Backend) GetChallenge(challengeId string) (*Challenge, error) {
	challenge := &Challenge{}
	err := backend.get("/v1/challenge/"+challengeId, challenge)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

func (backend *Backend) GetRandomChallenge() (*Challenge, error) {
	challenge := &Challenge{}
	err := backend.get("/v1/challenge/random", challenge)
	if err != nil {
		return nil, err
	}
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
