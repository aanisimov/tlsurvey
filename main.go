package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Variant struct {
	A string `json:"A"`
	B string `json:"B"`
	C string `json:"C"`
	D string `json:"D"`
	O string `json:"O"`
}

type Question struct {
	Index       int      `json:"index"`
	Description string   `json:"description"`
	Variants    Variant  `json:"variants"`
	Answers     []Answer `json:"answers"`
}

type Answer struct {
	Role        string `json:"role"`
	Team        string `json:"team"`
	Fio         string `json:"fio"`
	Timest      string `json:"timest"`
	Answer      string `json:"answer"`
	AnswerLabel string `json:"answerLabel"`
}

type Survey struct {
	ID           string     `json:"id"`
	Questions    []Question `json:"questions"`
	Participants []string   `json:"participants"`
}

func checkQuestionPresent(question Question, survey Survey) bool {
	for _, q := range survey.Questions {
		if q.Description == question.Description {
			return true
		}
	}
	return false
}

func parseCSV(filename string, survey Survey) Survey {
	lines, err := ReadCsv(filename)
	if err != nil {
		panic(err)
	}

	for idx, item := range lines[0] {
		question := Question{Description: item, Index: idx}
		if !checkQuestionPresent(question, survey) {
			survey.Questions = append(survey.Questions, question)
		}
	}

	questions := lines[0]

	for _, item := range lines[1:] {
		timestamp := item[0]
		role := item[1]
		team := item[2]
		fio := item[3]
		survey = addParticipant(survey, role, fio)
		for idx, column := range item {
			for idd, question := range survey.Questions {
				if question.Description == questions[idx] {
					ans := Answer{Role: role, Team: team, Fio: fio, Timest: timestamp, Answer: column}
					survey.Questions[idd].Answers = append(survey.Questions[idd].Answers, ans)
				}
			}
		}
	}

	return survey
}

func exportSurveyToJSON(survey Survey) error {
	file, _ := json.MarshalIndent(survey, "", " ")
	filename := fmt.Sprintf("%s.json", survey.ID)
	err := ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func importSurveyFromJSON(id string) {
	return
}

func getQuestionsFromCSV(questionsFileName string) []Question {
	lines, _ := ReadCsv(questionsFileName)
	questions := []Question{}
	for idx, item := range lines[0][1:] {
		variant := Variant{O: lines[1][idx+1], A: lines[2][idx+1], B: lines[3][idx+1], C: lines[4][idx+1], D: lines[5][idx+1]}
		question := Question{Index: idx + 1, Description: item, Variants: variant}
		questions = append(questions, question)
	}
	return questions
}

func getAnswerByFIO(question Question, fio string) Answer {
	for _, answer := range question.Answers {
		if answer.Fio == fio {
			return answer
		}
	}
	return Answer{Answer: "N/A", AnswerLabel: "N/A"}
}

func getQuestionCSVLine(question Question, survey Survey) []string {
	if question.Description == "Название команды" || question.Description == "Ваша роль:" || question.Description == "Ваше имя и фамилия" || question.Description == "Timestamp" || question.Description == "Отметка времени" {
		return nil
	}
	line := []string{question.Description}
	for _, p := range survey.Participants {
		fio := strings.Split(p, " (")[0]
		answer := ""
		if getAnswerByFIO(question, fio).AnswerLabel == "" {
			answer = getAnswerByFIO(question, fio).Answer
		} else {
			answer = getAnswerByFIO(question, fio).AnswerLabel
		}
		line = append(line, answer)
	}
	return line
}

func assignAnswersToLabel(question Question) Question {
	for idx, answer := range question.Answers {
		switch answer.Answer {
		case question.Variants.O:
			question.Answers[idx].AnswerLabel = "O"
		case question.Variants.A:
			question.Answers[idx].AnswerLabel = "A"
		case question.Variants.B:
			question.Answers[idx].AnswerLabel = "B"
		case question.Variants.C:
			question.Answers[idx].AnswerLabel = "C"
		case question.Variants.D:
			question.Answers[idx].AnswerLabel = "D"
		}
	}
	return question
}

func exportResultsToCSV(survey Survey) error {
	file, err := os.Create(fmt.Sprintf("%s_results.csv", survey.ID))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	firstLine := []string{""}
	firstLine = append(firstLine, survey.Participants...)
	writer.Write(firstLine)
	for _, q := range survey.Questions {
		line := getQuestionCSVLine(q, survey)
		if line != nil {
			writer.Write(line)
		}
	}
	return nil
}

func showDetailedResults(id string) {
	return
}

func ReadCsv(filename string) ([][]string, error) {

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func addParticipant(survey Survey, role string, fio string) Survey {
	newParticipant := fmt.Sprintf("%s (%s)", fio, role)
	for _, participant := range survey.Participants {
		if participant == newParticipant {
			return survey
		}
	}
	survey.Participants = append(survey.Participants, newParticipant)
	return survey
}

func main() {
	csvDirPath := flag.String("csvPath", "csv", "Path to folder with CSV files with responses")
	questionsCSVPath := flag.String("qPath", "questions/questions_answers.csv", "Path to CSV file with questions and answers variants")
	surveyID := flag.String("surveyID", "Результаты", "ID текущего опросника")
	flag.Parse()

	questions := getQuestionsFromCSV(*questionsCSVPath)
	survey := Survey{ID: *surveyID, Questions: questions}

	files, err := ioutil.ReadDir(*csvDirPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		survey = parseCSV(fmt.Sprintf("%s/%s", *csvDirPath, f.Name()), survey)
		err := exportSurveyToJSON(survey)
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	for i, question := range survey.Questions {
		labeledQuestion := assignAnswersToLabel(question)
		survey.Questions[i] = labeledQuestion
	}

	err = exportSurveyToJSON(survey)
	if err != nil {
		fmt.Printf("%s", err)
	}

	exportResultsToCSV(survey)
}
