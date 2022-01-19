// install:
// go get -u github.com/ledongthuc/pdf

package main


import (
    "html/template"
    "net/http"
    "log"
    "strconv"
    // "path/filepath"
    // "fmt"
    "strings"
    "bytes"
    "github.com/ledongthuc/pdf"
    // tokenize "github.com/jdkato/prose/tokenize"
    // "github.com/euskadi31/go-tokenizer"
    "regexp"
)




type StudentGradeData struct {
    StudentName string
    Grade int
    ReceivedPoints int
    TotalPoints int
    Success bool
    PdfContent string
    PdfTokenized []string
}


type GradedPageFormData struct {
    StudentName string
}


func CheckErr(err error) {
   if err != nil {
      log.Fatal(err)
   }
}

func StrToInt(strVar string) int{
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
    // array := regexp.MustCompile("[\\:\\,\\.\\s]+").Split(word, -1)
    return array
}


func TokenizeToWords(text string) []string{

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
    return splitWord( strings.ToLower(text))
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
        } else{

            pdf_path := r.FormValue("pdf_path")
            pdf_content, err := ReadPdf(pdf_path)
            CheckErr(err)

            pdf_tokenized := TokenizeToWords(pdf_content)

            points:= StrToInt(r.FormValue("received_points"))
            total_points := 10 

            // grade := (float32(points) / float32(total_points) * 10 //as float
            grade := int(float32(points) / float32(total_points) * 10) //as int

            data := StudentGradeData{
                Success: true,
                StudentName:   r.FormValue("name"),
                Grade: grade,
                TotalPoints: total_points,
                ReceivedPoints: points,
                PdfContent: pdf_content,
                PdfTokenized: pdf_tokenized,
            }

            log.Print("data:")
            log.Print(data)
            

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