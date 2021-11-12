package main

import (
	"fmt"
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

func main() {
	// fmt.Printf("gocv version: %s\n", gocv.Version())
	// fmt.Printf("opencv lib version: %s\n", gocv.OpenCVVersion())
	origImg := gocv.IMRead("./img/IMG_4038.jpg", gocv.IMReadColor)
	// image2 := gocv.IMRead("./img/IMG_4038.jpg", gocv.IMReadGrayScale)
	blurredImg := gocv.NewMat()
	greyImg := gocv.NewMat()
	threshedImg := gocv.NewMat()
	smallSide := float64(3024)
	minContour := smallSide * .04
	minBound := smallSide * .03
	minMinRectSize := smallSide * 0.05
	maxMinRectSize := smallSide * 0.21
	minExtent := 0.55
	maxExtent := 0.90

	gocv.GaussianBlur(origImg, &blurredImg, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
	gocv.CvtColor(blurredImg, &greyImg, gocv.ColorBGRToGray)

	gocv.AdaptiveThreshold(greyImg, &threshedImg, 255, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinary, 7, 2)

	origImg.Size()
	// gocv.Threshold(image2, &image2, 150, 255, gocv.ThresholdBinary)
	contours := gocv.FindContours(threshedImg, gocv.RetrievalTree, gocv.ChainApproxSimple)
	// gocv.DrawContours(&origImg, contours, -1, color.RGBA{0, 255, 0, 0}, 30)
	fmt.Println(contours.Size())
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		if contour.Size() < int(minContour) {
			continue
		}

		rect := gocv.BoundingRect(contour)
		if rect.Max.X < int(minBound) && rect.Max.Y < int(minBound) {
			continue
		}

		minRect := gocv.MinAreaRect(contour)
		rectWidth := float64(minRect.BoundingRect.Size().X)
		rectHeight := float64(minRect.BoundingRect.Size().Y)
		// minRect.BoundingRect.Size().X

		if rectWidth < minMinRectSize || rectHeight < minMinRectSize {
			//purple - minrect too thin
			// cv.drawContours(canvas, contours, i, [128, 0, 128, 255], 1, cv.LINE_8);
			continue
		}

		if rectWidth > maxMinRectSize && rectHeight > maxMinRectSize {
			//blue - minrect too large
			// cv.drawContours(canvas, contours, i, [0, 0, 255, 255], 1, cv.LINE_8);
			continue
		}

		if !ratioFits(minRect) {
			continue
		}

		area := gocv.ContourArea(contour)
		shapeExtent := area / (rectWidth * rectHeight)

		if shapeExtent < minExtent {
			//orange - shape extent too small
			// cv.drawContours(canvas, contours, i, [255, 128, 0, 255], 1, cv.LINE_8);
			continue
		}

		if shapeExtent > maxExtent {
			//red - shape extent too big
			// cv.drawContours(canvas, contours, i, [255, 0, 0, 255], 1, cv.LINE_8);
			continue
		}

		shape := Shape{
			FullContour:   contour,
			ParentContour: growContour(contour, smallSide*.02),
			ChildContour:  growContour(contour, smallSide*-.01),
			Type:          findType(shapeExtent),
		}

		// for i, contour := range contours.ToPoints() {
		fmt.Println(contour.Size())

		gocv.DrawContours(&origImg, contours, i, color.RGBA{0, 255, 0, 0}, 30)
	}

	gocv.IMWrite("./img/rendered.jpg", origImg)

}
func ratioFits(rect gocv.RotatedRect) bool {
	minRatio := 1.3 // set to 1.0 to get back edge of top right card
	maxRatio := 3.2

	ratio := getGreaterAspectRatio(rect)

	return ratio > minRatio && ratio < maxRatio
}

func getGreaterAspectRatio(minRect gocv.RotatedRect) float64 {
	width := float64(minRect.BoundingRect.Size().X)
	height := float64(minRect.BoundingRect.Size().Y)

	if width > height {
		return (width / height)
	}
	return (height / width)
}

type Shape struct {
	FullContour   gocv.PointsVector
	ParentContour gocv.PointsVector
	ChildContour  gocv.PointsVector
	Type          string
}

func findType(extent float64) string {
	switch {
	case extent < 0.666:
		return "diamond"
	case extent < 0.81:
		return "squiggle"
	default:
		return "oval"
	}
}

func growContour(contour gocv.PointsVector, resize float64) gocv.PointsVector {
	cSize := contour.Size()
	newC := gocv.NewMatWithSize(cSize/2, 2, gocv.MatTypeCV32S)

	prevPoint := image.Point{contour.ToPoints()[0], contour.ToPoints()[1]}
	pointPoint := image.Point{contour.ToPoints()[2], contour.ToPoints()[3]}
	nextPoint := image.Point{contour.ToPoints()[4], contour.ToPoints()[5]}

	for i := 6; i < cSize+6; i += 2 {
		dist := getDistVec(prevPoint, nextPoint)
	}
}

func getDistVec(p0, p1 image.Point) image.Point {
	return image.Point{
		p1.X - p0.X,
		p1.Y - p0.Y,
	}
}
