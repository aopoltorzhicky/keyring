package main

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/aopoltorzhicky/keyring"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if len(os.Args) < 2 {
		log.Panic().Msg("you have to point a command")
	}

	command := os.Args[1]

	log.Print("Enter keyring password: ")
	keyringPassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Panic().Err(err).Msg("error during setting password")
	}

	if err := keyring.Create(keyringPassword); err != nil {
		log.Panic().Err(err).Msg("error during creating keyring")
	}
	service, err := read("Enter service: ")
	if err != nil {
		log.Panic().Err(err).Msg("error during getting service name")
	}

	username, err := read("Enter username: ")
	if err != nil {
		log.Panic().Err(err).Msg("error during getting user name")
	}

	switch command {
	case "get":
		password, err := keyring.Keys.Get(service, username)
		if err != nil {
			log.Panic().Err(err).Msg("error during getting password")
		}
		log.Print(password)
	case "set":
		log.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			log.Panic().Err(err).Msg("error during setting password")
		}
		if err := keyring.Keys.Set(service, username, string(bytePassword)); err != nil {
			log.Panic().Err(err).Msg("error during setting password")
		}
		log.Print("success")
	case "delete":
		if err := keyring.Keys.Delete(service, username); err != nil {
			log.Panic().Err(err).Msg("error during deleting password")
		}
		log.Print("success")
	case "help":
		log.Print("You can use next commands:")
		log.Print("  keyring get service username")
		log.Print("  keyring set service username password")
		log.Print("  keyring delete service username")
	default:
		log.Panic().Msgf("unknown command: %s", command)
	}
}

func read(question string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	log.Print(question)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	text = strings.ReplaceAll(text, "\n", "")
	return text, nil
}
