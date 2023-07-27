package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// https://abit.itmo.ru/rating/master/budget/7431
const URL = "https://abitlk.itmo.ru/api/v1/rating/master/budget?program_id=%d" // &manager_key=&sort=&showLosers=true

type ApiResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Result  struct {
		Direction struct {
			Name   string `json:"direction_title"`
			Quota  int    `json:"budget_min"`
			Target int    `json:"target_reception"`
		} `json:"direction"`
		Applicants []Applicant `json:"general_competition"`
		Timestamp  time.Time   `json:"update_time"`
	} `json:"result"`
}

type Applicant struct {
	DiplomaAverage float64 `json:"diploma_average"`
	Score          float64 `json:"total_scores"`
	Priority       int     `json:"priority"`
	Originals      bool    `json:"is_send_original"`
	Snils          string  `json:"snils"`
	Status         string  `json:"status"`
}

func main() {
	flag.Usage = func() {
		fmt.Println("ItmoAbit CLI (v.1)")
		fmt.Printf("Usage: %s [OPTIONS] [Снилc] [ID программы]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	program := *flag.Int("p", 7431, "ID программы")

	// fmt.Println(*program)
	// if program == "" {
	// 	fmt.Println("ID программы не указано")
	// 	os.Exit(1)
	// }

	flag.Parse()

	snils := flag.Arg(0)
	if snils == "" {
		fmt.Println("Снилс не указан")
		os.Exit(1)
	}

	resp, err := http.Get(fmt.Sprintf(URL, program))

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data ApiResponse
	if err := json.Unmarshal(body, &data); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	if !data.OK {
		fmt.Println("Ошибка. Возможно не правильно указан ID програмы")
		os.Exit(1)
	}

	fmt.Printf("Направление: %s\n", data.Result.Direction.Name)
	fmt.Printf("Бюджетные места: %v\n", data.Result.Direction.Quota)
	fmt.Printf("Целевая квота: %v\n", data.Result.Direction.Target)
	fmt.Printf("Данные от: %v\n", data.Result.Timestamp)

	applicants := data.Result.Applicants

	// Индекс текущего поступающего
	// Current applicant index
	var cai int
	// var currentApplicant Applicant
	for i, a := range applicants {
		if a.Snils == snils {
			cai = i
			// currentApplicant = a
		}
	}

	if cai == 0 {
		fmt.Println("-------------------------------------------------")
		fmt.Printf("Снилс %s не найден в списке поступающих\n", snils)
		os.Exit(1)
	}

	fmt.Println("-------------------------------------------------")
	fmt.Printf("Балл ВИ+ИД: %v\n", applicants[cai].Score)
	fmt.Printf("Средний балл: %.4f\n", applicants[cai].DiplomaAverage)
	fmt.Printf("Текущее место: %v\n", cai+1)
	fmt.Printf("Оригиналы: %v\n", func() string {
		if applicants[cai].Originals {
			return "Да"
		}
		return "Нет"
	}())

	// Количество поступающих с приоритетом 1 и оригиналами
	var p11 int
	// Количество поступающих с приоритетом 1
	var p10 int

	for i := 0; i < cai; i++ {
		if !(applicants[i].Priority > 1) {
			p10++
			if applicants[i].Originals {
				p11++
			}
		}
	}

	fmt.Println("-------------------------------------------------")
	fmt.Printf("Место относительно приоритета 1: %v\n", p10+1)
	fmt.Printf("Место относительно приоритета 1 и оригиналов: %v\n", p11+1)

	// Количество поступающих у которых средний балл выше,
	// но которые еще не прошли ВИ.
	var p20 int
	// Количество поступающих у которых средний балл выше,
	// но которые еще не прошли ВИ и выбрали приоритет 1
	var p21 int
	// Количество поступающих у которых средний балл выше,
	// но которые еще не прошли ВИ и выбрали приоритет 1 и подали оригиналы
	var p22 int
	for i := cai + 1; i < len(applicants); i++ {
		if applicants[i].Score < 10 {
			if applicants[i].DiplomaAverage > applicants[cai].DiplomaAverage || applicants[i].Score != 0 {
				p20++
				if !(applicants[i].Priority > 1) {
					p21++
					if applicants[i].Originals {
						p22++
					}
				}

			}
		}
	}

	fmt.Println("-------------------------------------------------")
	fmt.Printf("Место с учетом не сдавших ВИ (считаем их 100): %v\n", cai+1+p20)
	fmt.Printf("Место с учетом не сдавших ВИ c приоритетом 1: %v\n", cai+1+p21)
	fmt.Printf("Место с учетом не сдавших ВИ c приоритетом 1 и оригиналами: %v\n", cai+1+p22)

	var p30 int
	var p31 int
	for i, a := range applicants {
		if i == cai {
			p30++
			p31++
			break
		}

		if a.Status == "recommended" {
			p30++
			if a.Originals {
				p31++
			}
		}
	}

	fmt.Println("-------------------------------------------------")
	fmt.Printf("Место в списке рекомендованных к зачислению: %v\n", p30)
	fmt.Printf("Место в списке рекомендованных к зачислению (только оригиналы): %v\n", p31)

}

func prettyPrint(data any) string {
	s, _ := json.MarshalIndent(data, "", "\t")
	return string(s)
}
