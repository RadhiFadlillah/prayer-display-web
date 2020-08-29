package backend

import (
	"encoding/json"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	fp "path/filepath"
	"time"

	"github.com/RadhiFadlillah/go-prayer"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/julienschmidt/httprouter"
)

type event struct {
	Name  string `json:"name"`
	Time  int64  `json:"time"`
	Iqama int64  `json:"iqama,omitempty"`
}

type imageData struct {
	URL          string `json:"url"`
	MainColor    string `json:"mainColor"`
	HeaderMain   string `json:"headerMain"`
	HeaderAccent string `json:"headerAccent"`
	HeaderFont   string `json:"headerFont"`
	FooterMain   string `json:"footerMain"`
	FooterAccent string `json:"footerAccent"`
	FooterFont   string `json:"footerFont"`
}

// loadData is handler for /api/data
func loadData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Fetch necessary data
	times := getEventTimes()
	images, err := loadImages()
	checkError(err)

	// Encode to json
	data := struct {
		Events []event     `json:"events"`
		Images []imageData `json:"images"`
	}{times, images}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&data)
	checkError(err)
}

func getEventTimes() []event {
	calc := (&prayer.Calculator{
		Latitude:          -2.2307069,
		Longitude:         113.9301163,
		Elevation:         5,
		CalculationMethod: prayer.Kemenag,
		AsrConvention:     prayer.Shafii,
		PreciseToSeconds:  false,
		AngleCorrection: prayer.AngleCorrection{
			prayer.Fajr:    0.66667,
			prayer.Sunrise: -0.66667,
			prayer.Zuhr:    1,
			prayer.Asr:     0.66667,
			prayer.Maghrib: 0.75,
			prayer.Isha:    0.66667,
		},
	}).Init().SetDate(time.Now())

	fajr := calc.Calculate(prayer.Fajr)
	sunrise := calc.Calculate(prayer.Sunrise)
	zuhr := calc.Calculate(prayer.Zuhr)
	asr := calc.Calculate(prayer.Asr)
	maghrib := calc.Calculate(prayer.Maghrib)
	isha := calc.Calculate(prayer.Isha)
	nextFajr := fajr.AddDate(0, 0, 1)

	iqamahFajr := fajr.Add(20 * time.Minute)
	iqamahZuhr := zuhr.Add(15 * time.Minute)
	iqamahAsr := asr.Add(10 * time.Minute)
	iqamahMaghrib := maghrib.Add(10 * time.Minute)
	iqamahIsha := isha.Add(10 * time.Minute)

	return []event{
		createEvent("fajr", fajr, iqamahFajr),
		createEvent("sunrise", sunrise),
		createEvent("zuhr", zuhr, iqamahZuhr),
		createEvent("asr", asr, iqamahAsr),
		createEvent("maghrib", maghrib, iqamahMaghrib),
		createEvent("isha", isha, iqamahIsha),
		createEvent("nextFajr", nextFajr),
	}
}

func loadImages() ([]imageData, error) {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// Read `display` directory
	exeDir := fp.Dir(exePath)
	imageDir := fp.Join(exeDir, "display")
	files, err := ioutil.ReadDir(imageDir)
	if err != nil {
		return nil, err
	}

	// Only take jpg or png image
	listImage := []imageData{}
	for _, f := range files {
		fileName := f.Name()
		fileExt := fp.Ext(fileName)
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {
			continue
		}

		imgPath := fp.Join(imageDir, fileName)
		imgData, err := extractImageData(imgPath)
		if err != nil {
			return nil, err
		}

		listImage = append(listImage, imgData)
	}

	return listImage, nil
}

func extractImageData(imgPath string) (data imageData, err error) {
	// Open image
	img, err := imgio.Open(imgPath)
	if err != nil {
		return
	}

	// Resize image by half to make calculation faster
	img = transform.Resize(img,
		img.Bounds().Dx()/2,
		img.Bounds().Dy()/2,
		transform.NearestNeighbor)

	// Crop header and footer, both is third of image height
	imgBounds := img.Bounds()
	imgHeight := imgBounds.Dy()
	oneThirdHeight := imgHeight / 3

	header := transform.Crop(img, image.Rect(
		imgBounds.Min.X, imgBounds.Min.Y,
		imgBounds.Max.X, imgBounds.Min.Y+oneThirdHeight))

	footer := transform.Crop(img, image.Rect(
		imgBounds.Min.X, imgBounds.Max.Y-oneThirdHeight,
		imgBounds.Max.X, imgBounds.Max.Y))

	// Get main color of image
	mainColor := getDominantColor(img)

	// Get color palette for header and footer
	hMain, hAccent, hFont := getColorPalette(header)
	fMain, fAccent, fFont := getColorPalette(footer)

	// Create URL
	imgURL := path.Join("/", "image", fp.Base(imgPath))

	// Create final data
	data = imageData{
		URL:          imgURL,
		MainColor:    mainColor.Hex(),
		HeaderMain:   colorToRGBA(hMain, 0.7),
		HeaderAccent: colorToRGBA(hAccent, 0.7),
		HeaderFont:   hFont.Hex(),
		FooterMain:   colorToRGBA(fMain, 0.7),
		FooterAccent: colorToRGBA(fAccent, 0.7),
		FooterFont:   fFont.Hex(),
	}

	return
}

func createEvent(name string, t time.Time, iqama ...time.Time) event {
	msTime := t.UnixNano() / int64(time.Millisecond)

	var msIqama int64
	if len(iqama) > 0 {
		msIqama = iqama[0].UnixNano() / int64(time.Millisecond)
	}

	return event{
		Name:  name,
		Time:  msTime,
		Iqama: msIqama,
	}
}
