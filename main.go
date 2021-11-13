package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

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

		contour = growContour(contour, smallSide*.02)
		// shape := Shape{
		// 	FullContour:   contour,
		// 	ParentContour: growContour(contour, smallSide*.02),
		// 	ChildContour:  growContour(contour, smallSide*-.01),
		// 	Type:          findType(shapeExtent),
		// }

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
	FullContour   gocv.PointVector
	ParentContour gocv.PointVector
	ChildContour  gocv.PointVector
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

func growContour(contour gocv.PointVector, resize float64) gocv.PointVector {
	return contour
	dataLength := contour.Size()
	points := contour.ToPoints()

	prevPoint := image.Point{points[0].X, points[0].Y}
	pointPoint := image.Point{points[1].X, points[1].Y}
	nextPoint := image.Point{points[2].X, points[2].Y}

	newC := gocv.NewMatWithSize(dataLength/2, 2, gocv.MatTypeCV32S)
	for i := 6; i < dataLength+6; i += 2 {
		dist := getDistVec(prevPoint, nextPoint)
		if dist.X != 0 || dist.Y != 0 {
			facing := getNormalOrtho(dist)
			// newC.SetIntAt(points[i%dataLength].X, dist.Y, int32(math.Floor(float64(pointPoint.X)+resize*float64(facing.X))))
			newC.SetIntAt(dist.X, dist.Y, int32(math.Floor(float64(pointPoint.X)+resize*float64(facing.X))))
			newC.SetIntAt(dist.X, dist.Y, int32(math.Floor(float64(pointPoint.X)+resize*float64(facing.X))))
			// gocv.NewMatWithSizeFromScalar()
			// gocv.NewMatFrom
		}
	}
	// return newC
	return gocv.NewPointVectorFromMat(newC)
}

// func getContourPoint(points []image.Point, index int) image.Point {
// 	return image.Point{
// 		X: points[index].X,
// 		Y: points[index+1].Y,
// 	}
// }

func getDistVec(p0, p1 image.Point) image.Point {
	return image.Point{
		p1.X - p0.X,
		p1.Y - p0.Y,
	}
}

func getNormalOrtho(dist image.Point) image.Point {
	length := math.Sqrt((float64(dist.X) * 2) + (float64(dist.Y) * 2))
	return image.Point{
		// the switching of this seems like a mistake
		X: dist.Y / int(length),
		Y: dist.X / int(length),
	}
}
