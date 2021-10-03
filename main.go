package main

import (
	"github.com/adlio/trello"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
	"time"
)

func main() {
	var debug bool
	var appKey string
	var token string
	var boardID string
	var listID string
	var maxAge time.Duration

	var rootCmd = &cobra.Command{
		Use:   "trello-helper",
		Short: "Automate common trello tasks",
	}

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logs")
	_ = rootCmd.PersistentFlags().MarkHidden("debug")
	rootCmd.PersistentFlags().StringVar(&appKey, "appKey", "", "trello app key")
	_ = rootCmd.MarkPersistentFlagRequired("appKey")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "trello authentication token")
	_ = rootCmd.MarkPersistentFlagRequired("token")

	var archiveOlderThanCmd = &cobra.Command{
		Use:   "archive-older-than",
		Short: "Archives all cards in a list with activity older than `maxAge`.",
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}

			client := trello.NewClient(appKey, token)

			ArchiveOlderThan(client, boardID, listID, maxAge)
		},
	}

	archiveOlderThanCmd.PersistentFlags().StringVar(&boardID, "board", "", "trello board ID")
	_ = archiveOlderThanCmd.MarkPersistentFlagRequired("board")
	archiveOlderThanCmd.PersistentFlags().StringVar(&listID, "list", "", "trello list ID")
	archiveOlderThanCmd.PersistentFlags().DurationVar(&maxAge, "maxAge", (90*24)*time.Hour, "trello authentication token")
	rootCmd.AddCommand(archiveOlderThanCmd)

	rootCmd.Execute()
}

func ArchiveOlderThan(client *trello.Client, boardID, listID string, maxAge time.Duration) {
	var err error
	var board *trello.Board
	var cards []*trello.Card

	board, err = client.GetBoard(boardID)
	if err != nil {
		logrus.Fatalf("failed to get board [%s]: %v", boardID, err)
	}

	cards, err = board.GetCards()

	var removed int
	var total int
	for _, card := range cards {
		var movedAt time.Time
		var actions trello.ActionCollection

		total += 1

		// filter cards not in the given list
		if listID != "" && card.IDList != listID {
			logrus.Debugf("skipping card due to list filter [%s]", card.ID)
			continue
		}

		actions, err = card.GetListChangeActions()
		if err != nil {
			logrus.Errorf("failed to get actions for card [%s]", card.ID)
		}

		actions = actions.FilterToListChangeActions()
		sort.Sort(sort.Reverse(actions))
		for i := range actions {
			movedAt = actions[i].Date
			break
		}

		// filter cards not yet expired
		if movedAt.IsZero() || time.Now().Before(movedAt.Add(maxAge)) {
			logrus.Debugf("skipping card due to last activity threshold [%s]", card.ID)
			continue
		}

		logrus.Debugf("archiving expired card [%s](%s)", card.ID, card.Name)
		removed += 1
		err = card.Archive()
		if err != nil {
			logrus.Errorf("failed to archive card [%s]: %v", card.ID, err)
		}
	}

	logrus.Infof("archived %d/%d cards", removed, total)
}
