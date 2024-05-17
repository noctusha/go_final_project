package repeatRule

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func NextDate(Now time.Time, Date string, Repeat string) (Alarm string, Err error) {

	/* if len(repeat) == 0 {
		// расписать вариант с удалением задачи
	} */

	correctDate, err := time.Parse("20060102", Date) // по умолчанию входить должно "date" из DB, в основной функции я временно назвал это datelinefromDB
	if err != nil {
		return "", errors.New("invalid date format (DB)") // здесь проверяем в верном ли формате была изначальная дата (это исходное время (из БД), от которого начинается отсчёт повторений)
	}

	if Repeat[0] != 'd' && Repeat[0] != 'y' && Repeat[0] != 'w' && Repeat[0] != 'm' {
		return "", errors.New("invalid repetition rate") // здесь первоначально проверяем на корректность поля repeat (из БД)
	}

	parts := strings.Split(Repeat, " ")

	if parts[0] == "y" { // вариант с годом
		if len(parts) != 1 {
			return "", errors.New("invalid repetition rate")
		}
		tmp := correctDate.AddDate(1, 0, 0)
		for tmp.Before(Now) {
			tmp = tmp.AddDate(1, 0, 0)
		}
		Alarm = tmp.Format("20060102")
	}

	if parts[0] == "d" { // вариант с днями
		if len(parts) != 2 {
			return "", errors.New("invalid repetition rate")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", err
		}

		if days > 400 {
			return "", errors.New("days value exceeded")
		}

		tmp := correctDate.AddDate(0, 0, days)
		for tmp.Before(Now) {
			tmp = tmp.AddDate(0, 0, days)
		}
		Alarm = tmp.Format("20060102")
	}

	if parts[0] == "w" {
		var day int
		if len(parts) != 2 {
			return "", errors.New("invalid repetition rate")
		}
		possibleWeekday := []int{}
		for i := 0; i < len(parts[1]); i += 2 {
			day, err = strconv.Atoi(string(parts[1][i]))
			if err != nil || day < 1 || day > 7 {
				return "", errors.New("invalid repetition rate")
			}
			possibleWeekday = append(possibleWeekday, day)
		}

		tmp := Now.AddDate(0, 0, 1)
	top:
		for {
			for i := 0; i < len(possibleWeekday); i++ {
				if int(tmp.Weekday()) == possibleWeekday[i] || int(tmp.Weekday())+7 == possibleWeekday[i] {
					Alarm = tmp.Format("20060102")
					break top
				}
			}
			tmp = tmp.AddDate(0, 0, 1)
		}
	}

	if parts[0] == "m" {
		if len(parts) == 1 || len(parts) > 3 {
			return "", errors.New("invalid repetition rate")
		}

		if len(parts) == 2 {
			var day int
			possibleDaystr := strings.Split(parts[1], ",")
			possibleDay := make([]int, 0, len(possibleDaystr))

			for i := 0; i < len(possibleDaystr); i++ {
				day, err = strconv.Atoi(possibleDaystr[i])
				if err != nil || day < -2 || day == 0 || day > 31 {
					return "", errors.New("invalid repetition rate")
				}
				possibleDay = append(possibleDay, day)
			}

			tmp := Now.AddDate(0, 0, 1)
			if Now.Before(correctDate) {
				tmp = correctDate.AddDate(0, 0, 1)
			}
			tmpPlusOne := tmp.AddDate(0, 0, 1)
			tmpPlusTwo := tmp.AddDate(0, 0, 2)
		topp:
			for {
				for i := 0; i < len(possibleDay); i++ {
					if int(tmp.Day()) == possibleDay[i] || (possibleDay[i] == -1 && int(tmpPlusOne.Day()) == 1) || (possibleDay[i] == -2 && int(tmpPlusTwo.Day()) == 1) {
						Alarm = tmp.Format("20060102")
						break topp
					}
				}
				tmp = tmp.AddDate(0, 0, 1)
				tmpPlusOne = tmpPlusOne.AddDate(0, 0, 1)
				tmpPlusTwo = tmpPlusTwo.AddDate(0, 0, 1)
			}
		}

		if len(parts) == 3 {
			var day int
			possibleDaystr := strings.Split(parts[1], ",")
			possibleDay := make([]int, 0, len(possibleDaystr))

			for i := 0; i < len(possibleDaystr); i++ {
				day, err = strconv.Atoi(possibleDaystr[i])
				if err != nil || day < -2 || day == 0 || day > 31 {
					return "", errors.New("invalid repetition rate")
				}
				possibleDay = append(possibleDay, day)
			}

			var month int
			possibleMonthstr := strings.Split(parts[2], ",")
			possibleMonth := make([]int, 0, len(possibleMonthstr))

			for i := 0; i < len(possibleMonthstr); i++ {
				month, err = strconv.Atoi(possibleMonthstr[i])
				if err != nil || month < 1 || month > 12 {
					return "", errors.New("invalid repetition rate")
				}
				possibleMonth = append(possibleMonth, month)
			}

			tmp := Now.AddDate(0, 0, 1)
			if Now.Before(correctDate) {
				tmp = correctDate.AddDate(0, 0, 1)
			}

		to:
			for {
				for i := 0; i < len(possibleMonth); i++ {
					if int(tmp.Month()) == possibleMonth[i] {
						tmpPlusOne := tmp.AddDate(0, 0, 1)
						tmpPlusTwo := tmp.AddDate(0, 0, 2)

						for {
							for i := 0; i < len(possibleDay); i++ {
								if int(tmp.Day()) == possibleDay[i] || (possibleDay[i] == -1 && int(tmpPlusOne.Day()) == 1) || (possibleDay[i] == -2 && int(tmpPlusTwo.Day()) == 1) {
									Alarm = tmp.Format("20060102")
									break to
								}
							}
							tmp = tmp.AddDate(0, 0, 1)
							tmpPlusOne = tmpPlusOne.AddDate(0, 0, 1)
							tmpPlusTwo = tmpPlusTwo.AddDate(0, 0, 1)
						}
					}
					tmp = tmp.AddDate(0, 0, 1)
				}
			}
		}
	}
	return Alarm, nil
}
