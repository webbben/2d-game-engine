package dialogv2

import (
	"strings"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

func InsertDialogVariables(sourceText string, playerInfo defs.PlayerInfo, dataman *datamanager.DataManager) string {
	if strings.Contains(sourceText, "{") {
		for _, v := range AllDialogVariables {
			insertString := ""

			switch v {
			case VarPlayerName:
				insertString = playerInfo.PlayerName
			case VarPlayerCulture:
				cultureDef := dataman.GetCultureDef(playerInfo.PlayerCulture)
				insertString = cultureDef.DisplayName
			default:
				logz.Panicln("InsertDialogVariables", "variable name not recognized:", v)
			}

			sourceText = strings.ReplaceAll(sourceText, v, insertString)
		}
	}

	return sourceText
}
