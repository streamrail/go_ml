package ml

import (
	"fmt"
	"github.com/alonsovidales/go_matrix"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Regression Linear and logistic regression structure
type Regression struct {
	X [][]float64 // Training set of values for each feature, the first dimension are the test cases
	Y []float64   // The training set with values to be predicted
	// 1st dim -> layer, 2nd dim -> neuron, 3rd dim theta
	Theta     []float64
	LinearReg bool // true indicates that this is a linear regression problem, false a logistic regression one
}

// CostFunction Calcualtes the cost function for the training set stored in the
// X and Y properties of the instance, and with the theta configuration.
// The lambda parameter controls the degree of regularization (0 means
// no-regularization, infinity means ignoring all input variables because all
// coefficients of them will be zero)
// The calcGrad param in case of true calculates the gradient in addition of the
// cost, and in case of false, only calculates the cost
func (rg *Regression) CostFunction(lambda float64, calcGrad bool) (j float64, grad [][][]float64, err error) {
	if len(rg.Y) != len(rg.X) {
		err = fmt.Errorf(
			"the number of test cases (X) %d doesn't corresponds with the number of values (Y) %d",
			len(rg.X),
			len(rg.Y))
		return
	}

	if len(rg.Theta) != len(rg.X[0]) {
		err = fmt.Errorf(
			"the Theta arg has a lenght of %d and the input data %d",
			len(rg.Theta),
			len(rg.X[0]))
		return
	}

	if rg.LinearReg {
		j, grad = rg.linearRegCostFunction(lambda, calcGrad)
	} else {
		j, grad = rg.logisticRegCostFunction(lambda, calcGrad)
	}

	return
}

// InitializeTheta Initialize the Theta property to an array of zeros with the
// lenght of the number of features on the X property
func (rg *Regression) InitializeTheta() {
	rand.Seed(int64(time.Now().Nanosecond()))
	rg.Theta = make([]float64, len(rg.X[0]))
}

// LinearHipotesis Returns the hipotesis result for Linear Regression algorithm
// for the thetas in the instance and the specified parameters
func (rg *Regression) LinearHipotesis(x []float64) (r float64) {
	for i := 0; i < len(x); i++ {
		r += x[i] * rg.Theta[i]
	}

	return
}

// linearRegCostFunction returns the cost and gradient for the current instance
// configuration
func (rg *Regression) linearRegCostFunction(lambda float64, calcGrad bool) (j float64, grad [][][]float64) {

	auxTheta := make([]float64, len(rg.Theta))
	copy(auxTheta, rg.Theta)
	theta := [][]float64{auxTheta}

	m := float64(len(rg.X))
	y := [][]float64{rg.Y}

	pred := mt.Trans(mt.Mult(rg.X, mt.Trans(theta)))
	errors := mt.SumAll(mt.Apply(mt.Sub(pred, y), powTwo)) / (2 * m)
	regTerm := (lambda / (2 * m)) * mt.SumAll(mt.Apply([][]float64{rg.Theta[1:]}, powTwo))

	j = errors + regTerm
	grad = [][][]float64{mt.Sum(mt.MultBy(mt.Mult(mt.Sub(pred, y), rg.X), 1/m), mt.MultBy(theta, lambda/m))}

	return
}

// LoadFile loads information from the local file located at filePath, and after
// parse it, returns the Regression ready to be used with all the information
// loaded
// The file format is:
//      X11 X12 ... X1N Y1
//      X21 X22 ... X2N Y2
//      ... ... ... ... ..
//      XN1 XN2 ... XNN YN
//
// Note: Use a single space as separator
func LoadFile(filePath string) (rg *Regression) {
	strInfo, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	rg = new(Regression)

	trainingData := strings.Split(string(strInfo), "\n")
	for _, line := range trainingData {
		if line == "" {
			break
		}

		var values []float64
		for _, value := range strings.Split(line, " ") {
			floatVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				panic(err)
			}
			values = append(values, floatVal)
		}
		rg.X = append(rg.X, values[:len(values)-1])
		rg.Y = append(rg.Y, values[len(values)-1])
	}

	return
}

// LogisticHipotesis returns the hipotesis result for Logistic Regression for
// the thetas in the instance and the specified parameters
func (rg *Regression) LogisticHipotesis(x []float64) (r float64) {
	for i := 0; i < len(x); i++ {
		r += x[i] * rg.Theta[i]
	}
	r = sigmoid(r)

	return
}

// logisticRegCostFunction returns the cost and gradient for the current
// instance configuration
func (rg *Regression) logisticRegCostFunction(lambda float64, calcGrad bool) (j float64, grad [][][]float64) {

	auxTheta := make([]float64, len(rg.Theta))
	copy(auxTheta, rg.Theta)
	theta := [][]float64{auxTheta}

	m := float64(len(rg.X))
	y := [][]float64{rg.Y}

	hx := mt.Apply(mt.Mult(theta, mt.Trans(rg.X)), sigmoid)
	j = (mt.Mult(mt.Apply(y, neg), mt.Trans(mt.Apply(hx, math.Log)))[0][0] -
		mt.Mult(mt.Apply(y, oneMinus), mt.Trans(mt.Apply(mt.Apply(hx, oneMinus), math.Log)))[0][0]) / m

	// Regularization
	theta[0][0] = 0
	j += lambda / (2 * m) * mt.SumAll(mt.Apply(theta, powTwo))

	// Gradient calc
	gradAux := mt.MultBy(mt.Mult(mt.Sub(hx, y), rg.X), 1/m)
	grad = [][][]float64{mt.Sum(gradAux, mt.MultBy(theta, lambda/m))}

	return
}

func (rg *Regression) Accuracy() float64 {
	m := len(rg.X)
	correct := 0.0

	for i := 0; i < m; i++ {
		x := rg.X[i]
		y := rg.Y[i]
		h := rg.LogisticHipotesis(x)

		if h >= 0.5 && y == 1 {
			correct++
		}
	}

	return correct / float64(m)
}

// MinimizeCost this metod splits the given data in three sets: training, cross
// validation, test. In order to calculate the optimal theta, tries with
// different possibilities and the training data, and check the best match with
// the cross validations, after obtain the best lambda, check the perfomand
// against the test set of data
func (rg *Regression) MinimizeCost(maxIters int, suffleData bool, verbose bool) (finalCost float64, trainingData *Regression, lambda float64, testData *Regression) {
	lambdas := []float64{0.0, 0.001, 0.003, 0.01, 0.03, 0.1, 0.3, 1, 3, 10, 30, 100, 300}

	if suffleData {
		rg = rg.shuffle()
	}

	// Get the 60% of the data as training data, 20% as cross validation, and
	// the remaining 20% as test data

	trainingData = &Regression{
		X:         rg.X,
		Y:         rg.Y,
		Theta:     rg.Theta,
		LinearReg: rg.LinearReg,
	}

	// Launch a process for each lambda in order to obtain the one with best
	// performance
	bestJ := math.Inf(1)
	bestA := 0.0
	bestLambda := 0.0
	initTheta := make([]float64, len(trainingData.Theta))
	copy(initTheta, trainingData.Theta)

	for _, posLambda := range lambdas {
		if verbose {
			fmt.Println("Checking Lambda:", posLambda)
		}
		copy(trainingData.Theta, initTheta)
		Fmincg(trainingData, posLambda, 10, verbose)

		j, _, _ := trainingData.CostFunction(posLambda, false)

		if bestJ > j {
			bestJ = j
			// bestLambda = posLambda
		}

		accuracy := trainingData.Accuracy()
		if accuracy > bestA {
			bestA = accuracy
			bestLambda = posLambda
		}
	}

	// Include the cross validation cases into the training for the final train
	Fmincg(trainingData, bestLambda, maxIters, verbose)

	rg.Theta = trainingData.Theta

	finalCost, _, _ = trainingData.CostFunction(bestLambda, false)
	bestLambda = bestLambda

	return
}

func (rg *Regression) getTheta() [][][]float64 {
	return [][][]float64{
		[][]float64{
			rg.Theta,
		},
	}
}

// rollThetasGrad retuns a 1xn array that will contian the theta of the instance
func (rg *Regression) rollThetasGrad(x [][][]float64) [][]float64 {
	return x[0]
}

func (rg *Regression) setTheta(t [][][]float64) {
	rg.Theta = t[0][0]
}

// shuffle redistribute randomly all the X and Y rows of the instance
func (rg *Regression) shuffle() (shuffledData *Regression) {
	rand.Seed(int64(time.Now().Nanosecond()))

	shuffledData = &Regression{
		X: make([][]float64, len(rg.X)),
		Y: make([]float64, len(rg.Y)),
	}

	for i, v := range rand.Perm(len(rg.X)) {
		shuffledData.X[v] = rg.X[i]
		shuffledData.Y[v] = rg.Y[i]
	}

	shuffledData.Theta = rg.Theta

	return
}

// unrollThetasGrad converts the given two dim slice to a tree dim slice in order
// to be used with the Fmincg function
func (rg *Regression) unrollThetasGrad(x [][]float64) [][][]float64 {
	return [][][]float64{
		x,
	}
}
