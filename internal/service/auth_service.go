package service

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/faikbairamov/soccer-manager/internal/auth"
	"github.com/faikbairamov/soccer-manager/internal/config"
	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var firstNames = []string{
	"Luca", "Marco", "James", "Carlos", "Ahmed", "David", "Kenji", "Omar",
	"Rafael", "Ivan", "Bruno", "Mateo", "Noah", "Elias", "Hugo", "Felix",
	"Leon", "Sven", "Andre", "Diego", "Niko", "Arda", "Lamine", "Harry",
	"Cristiano", "Lionel", "Kylian", "Erling", "Vinícius", "Giorgi",
	"Khvicha", "Jaba", "Levan", "Tornike", "Giorgi", "Antoine", "Pedri",
}

var lastNames = []string{
	"Silva", "Müller", "Garcia", "Rossi", "Smith", "Chen", "Tanaka", "Diallo",
	"Fernandez", "Kovač", "Petrov", "Okafor", "Berg", "Santos", "Schneider",
	"Güler", "Yamal", "Kane", "Ronaldo", "Messi", "Mbappé", "Haaland",
	"Júnior", "Kvaratskhelia", "Mikautadze", "Davitashvili", "Lochoshvili",
	"Griezmann", "Costa", "Martins", "Fischer", "Weber", "Park",
}

var countries = []string{
	"Brazil", "Germany", "Spain", "Italy", "France", "Portugal", "Argentina",
	"England", "Netherlands", "Croatia", "Japan", "Nigeria", "Senegal",
	"Sweden", "Belgium", "Uruguay", "Colombia", "Mexico", "South Korea",
	"Georgia", "Turkey", "Norway", "Morocco", "Ghana", "Austria",
}

var teamNames = []string{
	"Real Madrid", "Barcelona", "Manchester City", "Manchester United", "Liverpool",
	"Arsenal", "Chelsea", "Bayern Munich", "Borussia Dortmund", "Paris Saint-Germain",
	"Juventus", "AC Milan", "Inter Milan", "Napoli", "Atletico Madrid",
	"Sevilla", "Ajax", "Porto", "Benfica", "Sporting CP",
	"Dinamo Tbilisi", "Lokomotivi Tbilisi", "Torpedo, Kutaisi",
	"Meshakhte Tkibuli", "Dila Gori", "Iveria Khashuri",
}

func randomFirstName() string { return firstNames[rand.IntN(len(firstNames))] }
func randomLastName() string  { return lastNames[rand.IntN(len(lastNames))] }
func randomCountry() string   { return countries[rand.IntN(len(countries))] }
func randomTeamName() string  { return teamNames[rand.IntN(len(teamNames))] }

type AuthService struct {
	store repository.Storer
	cfg   *config.Config
}

func NewAuthService(store repository.Storer, cfg *config.Config) *AuthService {
	return &AuthService{store: store, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	existing, err := s.store.GetUserByEmail(ctx, email)
	if err == nil && existing.ID.Valid {
		return "", domain.ErrConflict
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}
	var token string
	err = s.store.WithTx(ctx, func(q repository.Querier) error {
		user, err := q.CreateUser(ctx, repository.CreateUserParams{Email: email, Password: hash})
		if err != nil {
			return err
		}
		team, err := q.CreateTeam(ctx, repository.CreateTeamParams{
			UserID:  user.ID,
			Name:    randomTeamName(),
			Country: randomCountry(),
		})
		if err != nil {
			return err
		}
		if err := createInitialPlayers(ctx, q, team.ID); err != nil {
			return err
		}
		token, err = auth.GenerateToken(user.ID.Bytes, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
		return err
	})
	return token, err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrUnauthorized
		}
		return "", err
	}

	if err := auth.CheckPassword(password, user.Password); err != nil {
		return "", domain.ErrUnauthorized
	}

	return auth.GenerateToken(user.ID.Bytes, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
}

func createInitialPlayers(ctx context.Context, q repository.Querier, teamID pgtype.UUID) error {
	positions := []struct {
		pos   repository.PlayerPosition
		count int
	}{
		{repository.PlayerPositionGoalkeeper, 3},
		{repository.PlayerPositionDefender, 6},
		{repository.PlayerPositionMidfielder, 6},
		{repository.PlayerPositionAttacker, 5},
	}

	for _, p := range positions {
		for range p.count {
			_, err := q.CreatePlayer(ctx, repository.CreatePlayerParams{
				TeamID:    teamID,
				FirstName: randomFirstName(),
				LastName:  randomLastName(),
				Country:   randomCountry(),
				Position:  p.pos,
				Age:       int32(rand.IntN(23) + 18),
				Value:     1_000_000,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
