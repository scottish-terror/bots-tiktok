package tiktokmod

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
)

// Holiday - Struct for Holiday data
type Holiday struct {
	ID      int
	Name    string
	Day     time.Time
	Message string
}

// UserData - Matrix of user accounts
type UserData struct {
	ID      int
	Name    string
	SlackID string
	Trello  string
	Github  string
	Email   string
}

// SprintData - DB Storage struct
type SprintData struct {
	V2ID        int
	TeamID      string
	SprintStart time.Time
	Duration    int
	RetroID     string
	SprintName  string
	WorkingDays int
}

// BugLabel - Bug Label Information
type BugLabel struct {
	ID       int
	BoardID  string
	BugLevel string
	LabelID  string
}

// Squad - Squad Information
type Squad struct {
	ID        int
	BoardID   string
	Squadname string
	LabelID   string
	SquadPts  int
}

// Chapter - Chapter Information
type Chapter struct {
	ID            int
	BoardID       string
	ChapterName   string
	LabelID       string
	ChapterPoints int
	ChapterCount  int
}

// CardReportData - Card Information
type CardReportData struct {
	CardID           string
	CardTitle        string
	Points           int
	CardURL          string
	List             string
	StartedInWorking time.Time
	StartedInPR      time.Time
	EnteredDone      time.Time
	Owners           string
}

// SprintPointsBySquad - Sprint Points by Squad
type SprintPointsBySquad struct {
	SprintName   string
	SquadName    string
	SprintPoints int
}

// RetroStruct - struct of Retro Board UIDs
type RetroStruct struct {
	TeamID  string
	RetroID string
}

type peeps struct {
	ID     int
	Sprint string
	UserID int
	Squad  string
}

// TotalSprint - array of SprintPointsBySquad
type TotalSprint []SprintPointsBySquad

// Squads - array of squad
type Squads []Squad

// Chapters - array of Chapter
type Chapters []Chapter

// ConnectDB - establish gsql connection to db
func ConnectDB(tiktok *TikTokConf, dbName string) (db *sql.DB, status bool, err error) {

	if tiktok.Config.UseGCP {
		cfg := mysql.Cfg(tiktok.Config.SQLHost, tiktok.Config.DBUser, tiktok.Config.DBPassword)
		cfg.DBName = dbName
		cfg.AllowNativePasswords = tiktok.Config.AllowNativePasswords
		cfg.AllowCleartextPasswords = tiktok.Config.AllowCleartextPasswords
		cfg.AllowAllFiles = tiktok.Config.AllowAllFiles
		cfg.ParseTime = tiktok.Config.ParseTime

		db, err = mysql.DialCfg(cfg)
		if err != nil {
			errTrap(tiktok, "DB Connection Error: ", err)
			return db, false, err
		}

		return db, true, nil
	}

	myConn := tiktok.Config.DBUser + ":" + tiktok.Config.DBPassword + "@tcp(" + tiktok.Config.SQLHost + ":" + tiktok.Config.SQLPort + ")/" + dbName
	myParams := "?allowAllFiles=" + strconv.FormatBool(tiktok.Config.AllowAllFiles)
	myParams = myParams + "&allowCleartextPasswords=" + strconv.FormatBool(tiktok.Config.AllowCleartextPasswords)
	myParams = myParams + "&allowNativePasswords=" + strconv.FormatBool(tiktok.Config.AllowNativePasswords)
	myParams = myParams + "&parseTime=" + strconv.FormatBool(tiktok.Config.ParseTime)
	connString := myConn + myParams

	db, err = sql.Open("mysql", connString)
	if err != nil {
		errTrap(tiktok, "DB Connection Error: ", err)
		return db, false, err
	}

	return db, true, nil

}

// PutDBSprint - Put sprint data into DB
func PutDBSprint(tiktok *TikTokConf, sOpts SprintData) error {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if status {

		stmt, err := db.Prepare("INSERT tiktok_main SET teamid=?,sprintstart=?,duration=?,retroid=?,sprintname=?,workingdays=?")
		if err != nil {
			errTrap(tiktok, "SQL Error db.Prepare in `PutDBSprint` ", err)
			return err
		}

		_, err = stmt.Exec(sOpts.TeamID, sOpts.SprintStart, sOpts.Duration, sOpts.RetroID, sOpts.SprintName, sOpts.WorkingDays)
		if err != nil {
			errTrap(tiktok, "SQL Error stmt.Exec in `PutDBSprint`", err)
			return err
		}

		return nil
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	return err

}

// GetDBSprint - Get sprint data out of DB
func GetDBSprint(tiktok *TikTokConf, teamID string) (sOpts SprintData, err error) {
	var attachments Attachment

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		err := db.QueryRow("SELECT * FROM tiktok_main where teamid=? order by sprintstart desc limit 1", teamID).Scan(
			&sOpts.V2ID,
			&sOpts.TeamID,
			&sOpts.SprintStart,
			&sOpts.Duration,
			&sOpts.RetroID,
			&sOpts.SprintName,
			&sOpts.WorkingDays)
		switch {
		case err == sql.ErrNoRows:
			errTrap(tiktok, "No rows returned for db.QueryRow on "+teamID, err)
		case err != nil:
			errTrap(tiktok, "db.QueryRow error: ", err)
		default:
			return sOpts, nil
		}
		return sOpts, err
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	if tiktok.Config.LogToSlack {
		LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
	}
	return sOpts, err

}

// GetRetroID - Get all retro board IDs into one slice
func GetRetroID(tiktok *TikTokConf, teamID string) (retroStruct []RetroStruct, err error) {
	var attachments Attachment
	var tretro RetroStruct

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("select teamid,retroid from tiktok_main where teamid=?", teamID)
		if err != nil {
			errTrap(tiktok, "DB Query Error in `GetRetroID` in `sql.go`", err)
			return retroStruct, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&tretro.TeamID,
				&tretro.RetroID); err != nil {
				errTrap(tiktok, "DB rows.Scan Error in `GetRetroID` in `sql.go`", err)
				return retroStruct, err
			}

			retroStruct = append(retroStruct, tretro)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection in `GetRetroID` in `sql.go`, bailing out", tiktok, attachments)
		}
		return retroStruct, err
	}

	return retroStruct, nil
}

// GetDBSquads - get all squads and label IDs in db
func GetDBSquads(tiktok *TikTokConf, boardID string) (allSquads Squads, err error) {
	var attachments Attachment
	var tsquad Squad

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_squads where boardid=?", boardID)
		if err != nil {
			errTrap(tiktok, "DB Query Error", err)
			return allSquads, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&tsquad.ID,
				&tsquad.BoardID,
				&tsquad.Squadname,
				&tsquad.LabelID); err != nil {
				errTrap(tiktok, "DB Query Error", err)
				return allSquads, err

			}
			tsquad.SquadPts = 0

			allSquads = append(allSquads, tsquad)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return allSquads, err
	}

	return allSquads, nil
}

// GetDBChapters - get all chapters and label IDs in db
func GetDBChapters(tiktok *TikTokConf, boardID string) (allChapters Chapters, err error) {
	var attachments Attachment
	var tchapter Chapter

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_chapters where boardid=?", boardID)
		if err != nil {
			errTrap(tiktok, "DB Query Error in `GetDBChapters` in `sql.go`", err)
			return allChapters, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&tchapter.ID,
				&tchapter.BoardID,
				&tchapter.ChapterName,
				&tchapter.LabelID); err != nil {
				errTrap(tiktok, "rows.Scan DB error in `GetDbChapters` in `sql.go`", err)
				return allChapters, err

			}
			tchapter.ChapterPoints = 0
			tchapter.ChapterCount = 0

			allChapters = append(allChapters, tchapter)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed DB connection in `GetDbChapters` in `sql.go`, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB connection in `GetDbChapters` in `sql.go`, bailing out...", tiktok, attachments)
		}
		return allChapters, err
	}

	return allChapters, nil
}

// GetIgnoreLabels - get all label IDs that should be ignored for a board
func GetIgnoreLabels(tiktok *TikTokConf, boardID string) (ignoreLabels []string, err error) {
	var attachments Attachment
	var uid int
	var labelID string

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_label_ignore where boardid=?", boardID)
		if err != nil {
			errTrap(tiktok, "DB query Error", err)
			return ignoreLabels, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&uid,
				&boardID,
				&labelID); err != nil {
				errTrap(tiktok, "DB Query Error", err)
				return ignoreLabels, err

			}

			ignoreLabels = append(ignoreLabels, labelID)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return ignoreLabels, err
	}

	return ignoreLabels, nil
}

// LabelIgnore - add a label to the ignore table
func LabelIgnore(opts Config, tiktok *TikTokConf, labelID string) error {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if status {

		stmt, err := db.Prepare("INSERT tiktok_label_ignore SET boardid=?,labelid=?")
		if err != nil {
			errTrap(tiktok, "SQL Error in LabelIgnore", err)
			return err
		}

		_, err = stmt.Exec(opts.General.BoardID, labelID)
		if err != nil {
			errTrap(tiktok, "SQL Error in LabelIgnore", err)
			return err
		}

		return nil
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	return err

}

// GetUser - get a user from DB
func GetUser(tiktok *TikTokConf, myField string, mySearch string) (user UserData, err error) {

	var attachments Attachment

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_users where " + myField + "='" + mySearch + "'")
		if err != nil {
			errTrap(tiktok, "DB Query Error `db.Query` on tiktok_users in `GetUser`", err)
			return user, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&user.ID,
				&user.Name,
				&user.SlackID,
				&user.Trello,
				&user.Github,
				&user.Email); err != nil {
				errTrap(tiktok, "DB Query Error `rows.Next` on tiktok_users in `GetUser`", err)
				return user, err
			}
		}

	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return user, err
	}

	return user, nil

}

// GetDBUsers - get all users
func GetDBUsers(tiktok *TikTokConf) (users []UserData, err error) {
	var attachments Attachment
	var u UserData

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_users")
		if err != nil {
			errTrap(tiktok, "DB Query Error on tiktok_users in `GetUser`", err)
			return users, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&u.ID,
				&u.Name,
				&u.SlackID,
				&u.Trello,
				&u.Github,
				&u.Email); err != nil {
				errTrap(tiktok, "DB Query Error", err)
				return users, err

			}

			users = append(users, u)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return users, err
	}

	return users, nil
}

// AddDBUser - Put user data into DB
func AddDBUser(tiktok *TikTokConf, users UserData) bool {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL Error in AddDBUser", err)
		return false
	}

	if status {

		stmt, err := db.Prepare("INSERT tiktok_users SET name=?,slackid=?,trello=?,github=?,email=?")
		if err != nil {
			errTrap(tiktok, "SQL Error in AddDBUser", err)
			return false
		}

		_, err = stmt.Exec(users.Name, users.SlackID, users.Trello, users.Github, users.Email)
		if err != nil {
			errTrap(tiktok, "SQL Error in AddDBUser", err)
			return false
		}

		return true
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	return false

}

// zeroCardDataDB - drop data in carddata table
func zeroCardDataDB(tiktok *TikTokConf) error {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL error in `PutCardData`", err)
		return err
	}

	if status {
		stmt, err := db.Prepare("TRUNCATE TABLE tiktok_cardtracker")
		if err != nil {
			errTrap(tiktok, "SQL Error (db.Prepare) in zeroCardData on TRUNCATE", err)
			return err
		}

		_, err = stmt.Exec()
		if err != nil {
			errTrap(tiktok, "SQL Error (stmt.Exec) in zeroCardData on TRUNCATE", err)
			return err
		}

		return nil
	}

	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection in `zeroCardData` in `sql.go`, bailing out...")
	}

	return err

}

// PutCardData - put card data to DB instead of CSV
func PutCardData(tiktok *TikTokConf, allCardData CardReportData, teamID string) error {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL error in `PutCardData`", err)
		return err
	}

	if status {
		stmt, err := db.Prepare("INSERT tiktok_cardtracker SET cardid=?,cardtitle=?,points=?,cardurl=?,list=?,startedinworking=?,startedinpr=?,entereddone=?,owners=?,team=?")
		if err != nil {
			errTrap(tiktok, "SQL Error (db.Prepare) in `PutCardData`", err)
			return err
		}

		_, err = stmt.Exec(allCardData.CardID, allCardData.CardTitle, allCardData.Points, allCardData.CardURL, allCardData.List, allCardData.StartedInWorking, allCardData.StartedInPR, allCardData.EnteredDone, allCardData.Owners, teamID)
		if err != nil {
			errTrap(tiktok, "SQL Error (stmt.Exec) in `PutCardData`", err)
			return err
		}

		return nil
	}

	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection in `PutCardData` in `sql.go`, bailing out...")
	}

	return err

}

// PutThemeCount - Update board theme counts for reporting
func PutThemeCount(tiktok *TikTokConf, allTheme Themes, sOpts SprintData, teamID string) error {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL error in `PutThemeCount`", err)
		return err
	}

	if status {

		today := time.Now().Local()
		today.Format("2006-01-02 15:04:05")

		for _, z := range allTheme {
			stmt, err := db.Prepare("INSERT tiktok_theme_count SET countdate=?,team=?,sprintname=?,labelname=?,qty=?")
			if err != nil {
				errTrap(tiktok, "SQL error in `PutThemeCount`", err)
				return err
			}

			_, err = stmt.Exec(today, teamID, sOpts.SprintName, z.Name, z.Pts)
			if err != nil {
				errTrap(tiktok, "SQL error in `PutThemeCount`", err)
				return err
			}
		}
		return nil
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}

	return err
}

// GetPreviousSprintPoints - Retrieve Previous sprint data from CloudSQL
func GetPreviousSprintPoints(tiktok *TikTokConf, sprintname string) (totalSprint TotalSprint, err error) {

	var tempPoints SprintPointsBySquad
	var attachments Attachment

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_sprint_squad_points where LOWER(sprintname)=?", sprintname)
		if err != nil {
			errTrap(tiktok, "`GetPreviousSprintPoints` Function error: DB Query Error", err)
			return totalSprint, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&tempPoints.SprintName,
				&tempPoints.SquadName,
				&tempPoints.SprintPoints); err != nil {
				errTrap(tiktok, "`GetPreviousSprintPoints` Function error: DB Query Error", err)
				return totalSprint, err

			}

			totalSprint = append(totalSprint, tempPoints)
		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return totalSprint, err
	}

	return totalSprint, nil

}

// GetHoliday - Get List of Holidays in SQL DB.  If year is 0, all dates are returned
func GetHoliday(tiktok *TikTokConf, year string) (theHolidays []Holiday, err error) {

	var tempHoliday Holiday
	var attachments Attachment
	var rows *sql.Rows

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		if year == "0" {
			rows, err = db.Query("SELECT * FROM tiktok_holidays ORDER BY holiday")
		} else {
			rows, err = db.Query("SELECT * FROM tiktok_holidays where YEAR(holiday)=? ORDER BY holiday", year)
		}
		if err != nil {
			errTrap(tiktok, "`GetHoliday` Function error: DB Query Error", err)
			return theHolidays, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&tempHoliday.ID,
				&tempHoliday.Name,
				&tempHoliday.Day,
				&tempHoliday.Message); err != nil {
				errTrap(tiktok, "`GetHoliday` Function error: DB Query Error", err)
				return theHolidays, err

			}

			theHolidays = append(theHolidays, tempHoliday)
		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return theHolidays, err
	}

	return theHolidays, nil
}

// IsHoliday - Check for Holidays in SQL DB
func IsHoliday(tiktok *TikTokConf, checkDate time.Time) (isHoliday bool, holiday Holiday) {
	var attachments Attachment

	// checks for holidays in PST
	loc, err := time.LoadLocation("America/Tijuana")
	if err != nil {
		errTrap(tiktok, "TZ Data Error", err)
		return false, holiday
	}

	t := checkDate.In(loc)
	today := t.Format("2006-01-02")

	db, status, errdb := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		err := db.QueryRow("SELECT * FROM tiktok_holidays where holiday=? limit 1", today).Scan(
			&holiday.ID,
			&holiday.Name,
			&holiday.Day,
			&holiday.Message)
		switch {
		case err == sql.ErrNoRows:
			if tiktok.Config.DEBUG {
				fmt.Println("No rows returned for db.QueryRow on Holiday Check in `sql.go`")
			}
			return false, holiday
		case err != nil:
			errTrap(tiktok, "db.QueryRow error", err)
			return false, holiday
		default:
			return true, holiday
		}
	}

	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection to db in sql.go for holiday check, bailing out - " + errdb.Error())
	}
	if tiktok.Config.LogToSlack {
		LogToSlack("Failed DB Connection in `sql.go` for IsHoliday Func, bailing out - "+errdb.Error(), tiktok, attachments)
	}

	return false, holiday
}

// RecordSquadSprintData - Record points for sprint per squad
func RecordSquadSprintData(tiktok *TikTokConf, totalPoints Squads, sprintName string, nonPoints int) bool {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL Error in RecordSquadSprintData", err)
		return false
	}

	if status {

		stmt, err := db.Prepare("INSERT tiktok_sprint_squad_points SET sprintname=?,squadname=?,squadpoints=?")
		if err != nil {
			errTrap(tiktok, "SQL Error in `db.Prepare` func `RecordSquadSprintData`", err)
			return false
		}

		for _, s := range totalPoints {
			_, err = stmt.Exec(sprintName, s.Squadname, s.SquadPts)
			if err != nil {
				errTrap(tiktok, "SQL Error in `stmt.Exec` func `RecordSquadSprintData`", err)
				return false
			}
		}
		_, err = stmt.Exec(sprintName, "Non-Squad", nonPoints)
		if err != nil {
			errTrap(tiktok, "SQL Error in `stmt.Exec` for non-squad in func `RecordSquadSprintData`", err)
			return false
		}

		return true
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	return false

}

//DupeTable - Duplicates table inside CloudSQL DB
func DupeTable(tiktok *TikTokConf, newTableName string, existTableName string) error {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL Error in RecordSquadSprintData", err)
		return err
	}

	if status {
		stmt, err := db.Prepare("CREATE TABLE " + newTableName + " LIKE " + existTableName)
		if err != nil {
			errTrap(tiktok, "SQL Error (db.Prepare) in DupeTable on CREATE TABLE", err)
			return err
		}

		_, err = stmt.Exec()
		if err != nil {
			errTrap(tiktok, "SQL Error (stmt.Exec) in DupeTable on CREATE TABLE", err)
			return err
		}

		stmt, err = db.Prepare("INSERT INTO " + newTableName + " SELECT * FROM " + existTableName)
		if err != nil {
			errTrap(tiktok, "SQL Error (db.Prepare) in DupeTable on INSERT INTO", err)
			return err
		}

		_, err = stmt.Exec()
		if err != nil {
			errTrap(tiktok, "SQL Error (stmt.Exec) in DupeTable on INSERT INTO", err)
			return err
		}

		return nil
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	return err
}

// RecordChapterCount - Record points for sprint per squad
func RecordChapterCount(tiktok *TikTokConf, chapterName string, listName string, cardCount int, teamName string) bool {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)
	if err != nil {
		errTrap(tiktok, "SQL Error in RecordChapterCount", err)
		return false
	}

	if status {

		timeStamp := time.Now().Local()
		timeStamp.Format("2006-01-02")

		stmt, err := db.Prepare("INSERT tiktok_chapter_cards SET timestamp=?,chaptername=?,listname=?,cards=?,team=?")
		if err != nil {
			errTrap(tiktok, "SQL Error in `db.Prepare` func `RecordChapterCount`", err)
			return false
		}

		_, err = stmt.Exec(timeStamp, chapterName, listName, cardCount, teamName)
		if err != nil {
			errTrap(tiktok, "SQL Error in `stmt.Exec` func `RecordChapterCount`", err)
			return false
		}

		return true
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Failed connection, bailing out...")
	}
	return false

}

// GetBugID - get all Bug label IDs for a given board
func GetBugID(tiktok *TikTokConf, boardID string) (bugs []BugLabel, err error) {
	var attachments Attachment
	var temp BugLabel

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_bug_label where boardid=?", boardID)
		if err != nil {
			errTrap(tiktok, "DB query Error in `GetBugID` function in `sql.go`", err)
			return bugs, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&temp.ID,
				&temp.BoardID,
				&temp.BugLevel,
				&temp.LabelID); err != nil {
				errTrap(tiktok, "DB rows.Scan error in `GetBugID` function in `sql.go`", err)
				return bugs, err

			}

			bugs = append(bugs, temp)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return bugs, err
	}

	return bugs, nil
}

// GetSquadMembership - Get list of squads a user is part of in a given sprint
func GetSquadMembership(tiktok *TikTokConf, dbUserID int, sprintName string) (userList []string, err error) {

	db, status, err := ConnectDB(tiktok, tiktok.Config.SQLDBName)

	var peeps peeps
	var attachments Attachment

	if status {

		rows, err := db.Query("SELECT * FROM tiktok_squad_peeps where userID=? AND sprint=?", dbUserID, sprintName)
		if err != nil {
			errTrap(tiktok, "DB query Error in `GetSquadMembership` function in `sql.go`", err)
			return userList, err
		}

		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&peeps.ID,
				&peeps.Sprint,
				&peeps.Squad,
				&peeps.UserID); err != nil {
				errTrap(tiktok, "DB rows.Scan error in `GetSquadMembership` function in `sql.go`", err)
				return userList, err

			}

			userList = append(userList, peeps.Squad)

		}
	} else {
		if tiktok.Config.DEBUG {
			fmt.Println("Failed connection, bailing out...")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Failed DB Connection, bailing out", tiktok, attachments)
		}
		return userList, err
	}

	return userList, nil
}
