package backend

import (
	"encoding/json"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	fp "path/filepath"
	"time"

	"github.com/RadhiFadlillah/go-prayer"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/julienschmidt/httprouter"
)

type event struct {
	Name string `json:"name"`
	Time int64  `json:"time"`
}

type imageData struct {
	Path string `json:"path"`

	MainColor string `json:"mainColor"`

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
		Times  []event     `json:"times"`
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
		{Name: "fajr", Time: unixMilli(fajr)},
		{Name: "iqamahFajr", Time: unixMilli(iqamahFajr)},
		{Name: "sunrise", Time: unixMilli(sunrise)},
		{Name: "zuhr", Time: unixMilli(zuhr)},
		{Name: "iqamahZuhr", Time: unixMilli(iqamahZuhr)},
		{Name: "asr", Time: unixMilli(asr)},
		{Name: "iqamahAsr", Time: unixMilli(iqamahAsr)},
		{Name: "maghrib", Time: unixMilli(maghrib)},
		{Name: "iqamahMaghrib", Time: unixMilli(iqamahMaghrib)},
		{Name: "isha", Time: unixMilli(isha)},
		{Name: "iqamahIsha", Time: unixMilli(iqamahIsha)},
		{Name: "nextFajr", Time: unixMilli(nextFajr)},
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
		if fileExt != ".png" && fileExt != ".jpg" {
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

	// Create final data
	data = imageData{
		Path:         imgPath,
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
