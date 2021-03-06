package main

import (
	"bytes"
	// "fmt"
	"github.com/deckarep/golang-set"
	"github.com/jdkato/prose/v2"
	"github.com/ledongthuc/pdf"
	"html/template"
	"log"
	"net/http"
	// "reflect"
	"regexp"
	"strconv"
	"strings"
)

type StudentGradeData struct {
	StudentName          string
	Grade                int
	ReceivedPoints       int
	TotalPoints          int
	Success              bool
	PdfContent           string
	PdfTokenized         []string
	AllNounsReport       []string
	AllNounsRequirements []string
	JaccardSimNouns      float32
	NrNounsIntersect     int
	PercNounsFromReq     float32
}

type GradedPageFormData struct {
	StudentName string
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func StrToInt(strVar string) int {
	intVar, err := strconv.Atoi(strVar)

	CheckErr(err)

	return intVar
}

func ReadPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func splitWord(word string) []string {
	array := regexp.MustCompile("[\\:\\,\\.\\?\\!\\...\\-\\'\\(\\)\\[\\]\\{\\}\\=\\_\\\"\\\\“\\\\”\\//\\s]+").Split(word, -1)
	return array
}

func TokenizeToWords(text string) []string {

	// t := tokenizer.New()

	// pdf_tokenized:= make([]string, 0)

	// // // sentences := strings.Split(text, ".")
	// sentences := splitWord(text)

	// log.Print("_-------------------")

	// for _, element := range sentences {
	//     log.Print(element)
	//     pdf_tokenized = append(pdf_tokenized, element)
	// }

	// // tokens := t.Tokenize(sentences[0])

	// log.Print("_-------------------")
	// log.Print(pdf_tokenized)
	// // log.Print(tokens)

	// return tokens

	// // t := tokenize.NewTreebankWordTokenizer()

	// // pdf_sentences := tokenize.TextToWords(text)

	// // // pdf_tokenized:= make([]string, 0)

	// // // // for index, element := range pdf_sentences {
	// // for _, element := range pdf_sentences {

	// //     sentence_tokenized := t.Tokenize(element) // split text to words
	// //     log.Print(sentence_tokenized)
	// //     // pdf_tokenized = append(pdf_tokenized, sentence_tokenized...)

	// // }

	// // // // return pdf_tokenized

	// // log.Print("-------------------------------------\n")
	// // log.Print(pdf_sentences)

	// // // // countriesCpy := make([]string, len(pdf_tokenized))
	// // // // copy(countriesCpy, pdf_tokenized) //copies neededCountries to countriesCpy
	// // // // return countriesCpy
	// return pdf_tokenized
	// // return pdf_sentences
	return splitWord(strings.ToLower(text))
}

// func POS_tagging_text(text string) {
// 	doc, err := prose.NewDocument(text)
// 	CheckErr(err)

// 	for _, ent := range doc.Tokens() {
// 		fmt.Println(ent.Text, ent.Label, ent.Tag)
// 		// Go GPE
// 		// Google GPE
// 	}

// }

func filter_nouns(text string) []string {
	result := make([]string, 0)

	doc, err := prose.NewDocument(text)
	CheckErr(err)

	for _, ent := range doc.Tokens() {
		if ent.Tag == "NN" || ent.Tag == "NNP" {
			result = append(result, ent.Text)
		}
	}
	return result
}

func stringArrToSet(str []string) mapset.Set {
	set := mapset.NewSet()

	for _, i := range str {
		set.Add(i)
	}

	return set
}

func jaccard_similarity(words1 []string, words2 []string) (float32, int, int) {
	set1 := stringArrToSet(words1)
	set2 := stringArrToSet(words2)

	intersection := set1.Intersect(set2)
	union := set1.Union(set2)
	// fmt.Print("intersection cardinality:", intersection.Cardinality())

	return float32(intersection.Cardinality()) / float32(union.Cardinality()), intersection.Cardinality(), union.Cardinality()

	//     fmt.Println(textdistance.LevenshteinDistance(words1, words2))
	// 	fmt.Println(textdistance.DamerauLevenshteinDistance(s1, s2))
	// 	fmt.Println(textdistance.JaroDistance(s1, s2))
	// 	fmt.Println(textdistance.JaroWinklerDistance(s1, s2))
	// textdistance.Jac/

}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// cwd, err := os.Getwd()

	// utils.CheckErr(err)
	// CheckErr(err)

	// tmpl_form:= template.Must(template.ParseFiles(filepath.Join(cwd, "./templates/index.html")))

	tmpl_form := template.Must(template.ParseFiles("templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			tmpl_form.Execute(w, nil)
			return
		} else {
			requirements_path := r.FormValue("requirements_path")
			requirements_content, err := ReadPdf(requirements_path)
			CheckErr(err)
			requirements_tokenized := TokenizeToWords(requirements_content)
			requirements_tokenized_joined := strings.Join(requirements_tokenized[:], " ")
			all_nouns_requirements := filter_nouns(requirements_tokenized_joined)

			pdf_path := r.FormValue("pdf_path")

			pdf_content, err := ReadPdf(pdf_path)
			CheckErr(err)

			pdf_tokenized := TokenizeToWords(pdf_content)

			// POS_tagging_text(pdf_content)
			// all_nouns := filter_nouns(pdf_content)

			pdf_tokenized_joined := strings.Join(pdf_tokenized[:], " ")
			all_nouns_report := filter_nouns(pdf_tokenized_joined)

			jaccard_similarity_nouns, nr_nouns_intersect, _ := jaccard_similarity(all_nouns_requirements, all_nouns_report)
			nouns_set_req := stringArrToSet(all_nouns_requirements)
			log.Println("nr of nouns in the requirements:", float32(nouns_set_req.Cardinality()))
			log.Println("nr nouns intersect:", nr_nouns_intersect)
			perc_nouns_from_req := float32(nr_nouns_intersect) / float32(nouns_set_req.Cardinality())

			points := StrToInt(r.FormValue("received_points"))
			total_points := 10

			// grade := (float32(points) / float32(total_points) * 10 //as float
			grade := int(float32(points) / float32(total_points) * 10) //as int

			data := StudentGradeData{
				Success:              true,
				StudentName:          r.FormValue("name"),
				Grade:                grade,
				TotalPoints:          total_points,
				ReceivedPoints:       points,
				PdfContent:           pdf_content,
				PdfTokenized:         pdf_tokenized,
				AllNounsReport:       all_nouns_report,
				AllNounsRequirements: all_nouns_requirements,
				JaccardSimNouns:      jaccard_similarity_nouns,
				NrNounsIntersect:     nr_nouns_intersect,
				PercNounsFromReq:     perc_nouns_from_req,
			}

			// log.Print("data:")
			// log.Print(data)

			// tmpl_form.Execute(w, struct{ Success bool; Grade int}{true, 10})
			// tmpl_form.Execute(w, struct{ Success bool}{true})
			tmpl_form.Execute(w, data)
		}

	})

	// http.HandleFunc("/", indexHandler)

	////////////////// serve
	http.ListenAndServe(":8080", nil)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
