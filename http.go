package main

import (
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Data map[string]any

// Init initialize server router
func InitRouter() *http.ServeMux {
	mux := http.DefaultServeMux
	mux.Handle("/assets/", AssetsHandler("/assets/", Assets, "./static"))
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/download", downloadHandler)
	mux.HandleFunc("/delete", deleteHandler)
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fis := Range(upoladDir)
	d := Data{"files": fis}
	render(w, "index.html", &d)
}

// download file
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("id")
	path = joinPath(path)
	fd, err := os.Stat(path)
	if err != nil {
		Json(w, Data{"msg": fmt.Sprintf("cont find file %s.", path)})
		return
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(int(fd.Size())))
	w.Header().Set("Content-Disposition", "attachment;filename="+fd.Name())

	http.ServeFile(w, r, path)
	Debug("%s download %s", r.RemoteAddr, fd.Name())
}

// delete file
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("id")
	path = joinPath(path)
	fd, err := os.Stat(path)
	if err != nil {
		Json(w, Data{"msg": fmt.Sprintf("cont find file %s.", path)})
		return
	}
	os.Remove(path)
	Debug("%s delete   %s", r.RemoteAddr, fd.Name())
	redirect(w, r, "/")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		render(w, "upload.html", nil)
	} else {
		r.ParseMultipartForm(32 << 20) // 32M
		if r.MultipartForm == nil {
			return
		}

		for _, multifile := range r.MultipartForm.File {
			for _, fd := range multifile {
				// safe file.
				SaveFile(fd)
				Debug("%s upload   %s", r.RemoteAddr, fd.Filename)
			}
		}
		redirect(w, r, "/")
	}
}

func SaveFile(fd *multipart.FileHeader) error {
	path := joinPath(fd.Filename)

	// new file tmp
	if _, err := os.Stat(path); err != nil {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	// get tmp size
	fi, _ := os.Stat(path)

	// check if complime
	if fi.Size() == fd.Size {
		// upload finish.
		return nil
	}
	curSize := fi.Size()

	// open tmp file
	newFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer newFile.Close()

	// open upload file
	upFile, err := fd.Open()
	if err != nil {
		return err
	}
	defer upFile.Close()

	// seek
	upFile.Seek(curSize, 0)
	newFile.Seek(curSize, 0)

	bar := NewProcessBar(fd.Size, curSize)
	bar.dc = fd.Filename
	defer bar.Close()
	// save to tmp

	_, err = io.Copy(newFile, io.TeeReader(upFile, bar))

	// buf := make([]byte, 1024)
	// for {
	// 	n, err := io.TeeReader(upFile, bar).Read(buf)
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	_, err = newFile.Write(buf[:n])
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return err
}

type ProcessBar struct {
	dc    string
	Total int64
	Cur   int64
}

func NewProcessBar(max, cur int64) *ProcessBar {
	return &ProcessBar{Total: max, Cur: cur}
}

func (wc *ProcessBar) Write(p []byte) (int, error) {
	n := len(p)
	wc.Cur += int64(n)
	wc.render()
	return n, nil
}

func (wc *ProcessBar) Close() {
	fmt.Println("")
}

func (wc *ProcessBar) render() {
	cur := wc.Cur
	max := wc.Total

	present := float64(cur) / float64(max) * 100
	i := int(present / 4)
	if i > 25 {
		i = 25
	}
	h := strings.Repeat("▅", i) + strings.Repeat(" ", 25-i)

	fmt.Printf("\r%s", strings.Repeat(" ", 78))
	fmt.Printf("\r%-8s[%s]%3.0f%%",
		format(max), h, present)
}

func format(s int64) string {
	var tmp = []string{"B", "KB", "MB", "GB", "TB"}
	i, p, q := 0, 0.0, float64(s)
	for ; i < len(tmp); i++ {
		p = q / 1024
		if p < 1 {
			break
		}
		q = p
	}
	return fmt.Sprintf("%.2f%s", q, tmp[i])
}

type FileInfo struct {
	Name string // base name of the file
	Path string
	Size int64 // length in bytes for regular files; system-dependent for others
	//ModTime time.Time // modification time
	IsDir bool // abbreviation for Mode().IsDir()

}

func Range(dir string) (fis []FileInfo) {
	filepath.Walk(dir, func(path string, d fs.FileInfo, err error) error {
		if path == dir {
			return nil
		}
		if err != nil {
			return err
		}
		if d.IsDir() {
			// TODO add forder
			return nil
		}
		f := FileInfo{
			Name: d.Name(),
			Path: cleanPath(path),
			Size: d.Size(),
		}
		fis = append(fis, f)
		return nil
	})
	return
}

func cleanPath(path string) string {
	return strings.TrimPrefix(path, upoladDir)
}

func joinPath(path string) string {
	if filepath.HasPrefix(path, upoladDir) {
		return path
	}
	return filepath.Join(upoladDir, path)
}
