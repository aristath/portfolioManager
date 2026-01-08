package symbolic_regression

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// NodeType represents the type of a node in the formula tree
type NodeType int

const (
	NodeTypeConstant NodeType = iota
	NodeTypeVariable
	NodeTypeOperation
)

// Operation represents a mathematical operation
type Operation int

const (
	OpAdd Operation = iota
	OpSubtract
	OpMultiply
	OpDivide
	OpPower
	OpMax
	OpMin
	OpSqrt
	OpLog
	OpExp
	OpAbs
	OpNegate
)

// Node represents a node in the formula tree
type Node struct {
	Type     NodeType
	Value    float64   // For constants
	Variable string    // For variables
	Op       Operation // For operations
	Left     *Node     // Left child (for binary ops) or operand (for unary ops)
	Right    *Node     // Right child (for binary ops)
}

// Evaluate evaluates the formula tree with given variable values
func (n *Node) Evaluate(variables map[string]float64) float64 {
	switch n.Type {
	case NodeTypeConstant:
		return n.Value
	case NodeTypeVariable:
		if variables != nil {
			if val, ok := variables[n.Variable]; ok {
				return val
			}
		}
		return 0.0 // Default to 0 if variable not found
	case NodeTypeOperation:
		return n.evaluateOperation(variables)
	default:
		return 0.0
	}
}

// evaluateOperation evaluates an operation node
func (n *Node) evaluateOperation(variables map[string]float64) float64 {
	switch n.Op {
	case OpAdd:
		return n.Left.Evaluate(variables) + n.Right.Evaluate(variables)
	case OpSubtract:
		return n.Left.Evaluate(variables) - n.Right.Evaluate(variables)
	case OpMultiply:
		return n.Left.Evaluate(variables) * n.Right.Evaluate(variables)
	case OpDivide:
		right := n.Right.Evaluate(variables)
		if math.Abs(right) < 1e-10 {
			return 1.0 // Safe default for division by zero
		}
		return n.Left.Evaluate(variables) / right
	case OpPower:
		left := n.Left.Evaluate(variables)
		right := n.Right.Evaluate(variables)
		if left < 0 && right != math.Trunc(right) {
			// Negative base with non-integer exponent -> return safe value
			return 0.0
		}
		return math.Pow(left, right)
	case OpMax:
		left := n.Left.Evaluate(variables)
		right := n.Right.Evaluate(variables)
		return math.Max(left, right)
	case OpMin:
		left := n.Left.Evaluate(variables)
		right := n.Right.Evaluate(variables)
		return math.Min(left, right)
	case OpSqrt:
		val := n.Left.Evaluate(variables)
		if val < 0 {
			return 0.0 // Safe default for sqrt of negative
		}
		return math.Sqrt(val)
	case OpLog:
		val := n.Left.Evaluate(variables)
		if val <= 0 {
			return 0.0 // Safe default for log of non-positive
		}
		return math.Log(val)
	case OpExp:
		val := n.Left.Evaluate(variables)
		// Clamp to prevent overflow
		if val > 10 {
			val = 10
		} else if val < -10 {
			val = -10
		}
		return math.Exp(val)
	case OpAbs:
		return math.Abs(n.Left.Evaluate(variables))
	case OpNegate:
		return -n.Left.Evaluate(variables)
	default:
		return 0.0
	}
}

// String returns a string representation of the formula
func (n *Node) String() string {
	switch n.Type {
	case NodeTypeConstant:
		return formatFloat(n.Value)
	case NodeTypeVariable:
		return n.Variable
	case NodeTypeOperation:
		return n.operationString()
	default:
		return "0"
	}
}

// operationString returns string representation of an operation
func (n *Node) operationString() string {
	switch n.Op {
	case OpAdd:
		return "(" + n.Left.String() + " + " + n.Right.String() + ")"
	case OpSubtract:
		return "(" + n.Left.String() + " - " + n.Right.String() + ")"
	case OpMultiply:
		return "(" + n.Left.String() + " * " + n.Right.String() + ")"
	case OpDivide:
		return "(" + n.Left.String() + " / " + n.Right.String() + ")"
	case OpPower:
		return "pow(" + n.Left.String() + ", " + n.Right.String() + ")"
	case OpMax:
		return "max(" + n.Left.String() + ", " + n.Right.String() + ")"
	case OpMin:
		return "min(" + n.Left.String() + ", " + n.Right.String() + ")"
	case OpSqrt:
		return "sqrt(" + n.Left.String() + ")"
	case OpLog:
		return "log(" + n.Left.String() + ")"
	case OpExp:
		return "exp(" + n.Left.String() + ")"
	case OpAbs:
		return "abs(" + n.Left.String() + ")"
	case OpNegate:
		return "-(" + n.Left.String() + ")"
	default:
		return "0"
	}
}

// formatFloat formats a float64 to a readable string
func formatFloat(f float64) string {
	if f == math.Trunc(f) {
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.10f", f), "0"), ".")
	}
	return fmt.Sprintf("%.6f", f)
}

// RandomFormula generates a random formula tree
func RandomFormula(variables []string, maxDepth, maxNodes int) *Node {
	if len(variables) == 0 {
		// No variables, return constant
		return &Node{
			Type:  NodeTypeConstant,
			Value: rand.Float64()*2.0 - 1.0, // Random between -1 and 1
		}
	}

	return randomFormulaRecursive(variables, 0, maxDepth, maxNodes, 0)
}

// randomFormulaRecursive recursively generates a random formula
func randomFormulaRecursive(variables []string, depth, maxDepth, maxNodes, currentNodes int) *Node {
	if currentNodes >= maxNodes {
		// Force terminal node
		return randomTerminal(variables)
	}

	if depth >= maxDepth {
		// Max depth reached, return terminal
		return randomTerminal(variables)
	}

	// 30% chance of terminal at each level (increases with depth)
	terminalChance := 0.3 + float64(depth)*0.1
	if rand.Float64() < terminalChance {
		return randomTerminal(variables)
	}

	// Choose operation type
	opType := rand.Intn(2) // 0 = binary, 1 = unary

	if opType == 0 {
		// Binary operation
		binaryOps := []Operation{OpAdd, OpSubtract, OpMultiply, OpDivide, OpPower, OpMax, OpMin}
		op := binaryOps[rand.Intn(len(binaryOps))]

		return &Node{
			Type:  NodeTypeOperation,
			Op:    op,
			Left:  randomFormulaRecursive(variables, depth+1, maxDepth, maxNodes, currentNodes+1),
			Right: randomFormulaRecursive(variables, depth+1, maxDepth, maxNodes, currentNodes+2),
		}
	} else {
		// Unary operation
		unaryOps := []Operation{OpSqrt, OpLog, OpExp, OpAbs, OpNegate}
		op := unaryOps[rand.Intn(len(unaryOps))]

		return &Node{
			Type: NodeTypeOperation,
			Op:   op,
			Left: randomFormulaRecursive(variables, depth+1, maxDepth, maxNodes, currentNodes+1),
		}
	}
}

// randomTerminal generates a random terminal node (constant or variable)
func randomTerminal(variables []string) *Node {
	if len(variables) == 0 {
		return &Node{
			Type:  NodeTypeConstant,
			Value: rand.Float64()*2.0 - 1.0,
		}
	}

	// 50% chance of variable, 50% chance of constant
	if rand.Float64() < 0.5 {
		return &Node{
			Type:     NodeTypeVariable,
			Variable: variables[rand.Intn(len(variables))],
		}
	}

	return &Node{
		Type:  NodeTypeConstant,
		Value: rand.Float64()*2.0 - 1.0,
	}
}

// Mutate mutates a formula tree
func Mutate(formula *Node, variables []string, mutationRate float64) *Node {
	if rand.Float64() > mutationRate {
		return formula.Copy() // No mutation
	}

	return mutateRecursive(formula, variables, mutationRate)
}

// mutateRecursive recursively mutates a node
func mutateRecursive(node *Node, variables []string, mutationRate float64) *Node {
	if rand.Float64() > mutationRate {
		// Don't mutate this node, but may mutate children
		if node.Type == NodeTypeOperation {
			return &Node{
				Type:  NodeTypeOperation,
				Op:    node.Op,
				Left:  mutateRecursive(node.Left, variables, mutationRate),
				Right: node.Right,
			}
		}
		return node.Copy()
	}

	// Mutate this node
	mutationType := rand.Intn(4)

	switch mutationType {
	case 0:
		// Replace with random terminal
		return randomTerminal(variables)
	case 1:
		// Change constant value
		if node.Type == NodeTypeConstant {
			return &Node{
				Type:  NodeTypeConstant,
				Value: node.Value + (rand.Float64()*2.0-1.0)*0.1, // Small random change
			}
		}
		return randomTerminal(variables)
	case 2:
		// Change operation
		if node.Type == NodeTypeOperation {
			if node.Right != nil {
				// Binary op
				binaryOps := []Operation{OpAdd, OpSubtract, OpMultiply, OpDivide, OpPower, OpMax, OpMin}
				return &Node{
					Type:  NodeTypeOperation,
					Op:    binaryOps[rand.Intn(len(binaryOps))],
					Left:  node.Left.Copy(),
					Right: node.Right.Copy(),
				}
			} else {
				// Unary op
				unaryOps := []Operation{OpSqrt, OpLog, OpExp, OpAbs, OpNegate}
				return &Node{
					Type: NodeTypeOperation,
					Op:   unaryOps[rand.Intn(len(unaryOps))],
					Left: node.Left.Copy(),
				}
			}
		}
		return randomTerminal(variables)
	case 3:
		// Insert random subtree
		return RandomFormula(variables, 2, 5)
	default:
		return node.Copy()
	}
}

// Copy creates a deep copy of the node
func (n *Node) Copy() *Node {
	if n == nil {
		return nil
	}

	copy := &Node{
		Type:     n.Type,
		Value:    n.Value,
		Variable: n.Variable,
		Op:       n.Op,
	}

	if n.Left != nil {
		copy.Left = n.Left.Copy()
	}
	if n.Right != nil {
		copy.Right = n.Right.Copy()
	}

	return copy
}

// Crossover performs crossover between two formulas
func Crossover(formula1, formula2 *Node) (*Node, *Node) {
	// Find crossover points
	point1 := findRandomNode(formula1)
	point2 := findRandomNode(formula2)

	if point1 == nil || point2 == nil {
		// Can't crossover, return copies
		return formula1.Copy(), formula2.Copy()
	}

	// Swap subtrees
	child1 := formula1.Copy()
	child2 := formula2.Copy()

	// Replace point1 in child1 with point2 from formula2
	replaceNode(child1, point1, point2.Copy())
	// Replace point2 in child2 with point1 from formula1
	replaceNode(child2, point2, point1.Copy())

	return child1, child2
}

// findRandomNode finds a random node in the tree (for crossover)
func findRandomNode(node *Node) *Node {
	if node == nil {
		return nil
	}

	// Collect all nodes
	nodes := collectNodes(node)
	if len(nodes) == 0 {
		return nil
	}

	return nodes[rand.Intn(len(nodes))]
}

// collectNodes collects all nodes in the tree
func collectNodes(node *Node) []*Node {
	if node == nil {
		return nil
	}

	nodes := []*Node{node}

	if node.Left != nil {
		nodes = append(nodes, collectNodes(node.Left)...)
	}
	if node.Right != nil {
		nodes = append(nodes, collectNodes(node.Right)...)
	}

	return nodes
}

// replaceNode replaces a node in the tree (used for crossover)
func replaceNode(root, target, replacement *Node) bool {
	if root == nil {
		return false
	}

	if root == target {
		// Can't replace root, return false
		return false
	}

	if root.Left == target {
		root.Left = replacement
		return true
	}
	if root.Right == target {
		root.Right = replacement
		return true
	}

	if root.Left != nil && replaceNode(root.Left, target, replacement) {
		return true
	}
	if root.Right != nil && replaceNode(root.Right, target, replacement) {
		return true
	}

	return false
}

// FitnessType represents the type of fitness function
type FitnessType int

const (
	FitnessTypeMAE      FitnessType = iota // Mean Absolute Error (for expected returns)
	FitnessTypeRMSE                        // Root Mean Squared Error
	FitnessTypeSpearman                    // Spearman correlation (for ranking quality)
)

// CalculateFitness calculates fitness of a formula
func CalculateFitness(formula *Node, examples []TrainingExample, fitnessType FitnessType) float64 {
	if len(examples) == 0 {
		return math.MaxFloat64 // Worst possible fitness
	}

	switch fitnessType {
	case FitnessTypeMAE:
		return calculateMAE(formula, examples)
	case FitnessTypeRMSE:
		return calculateRMSE(formula, examples)
	case FitnessTypeSpearman:
		return calculateSpearman(formula, examples)
	default:
		return math.MaxFloat64
	}
}

// calculateMAE calculates Mean Absolute Error
func calculateMAE(formula *Node, examples []TrainingExample) float64 {
	sum := 0.0
	for _, ex := range examples {
		predicted := formula.Evaluate(getVariableMap(ex.Inputs))
		actual := ex.TargetReturn
		sum += math.Abs(predicted - actual)
	}
	return sum / float64(len(examples))
}

// calculateRMSE calculates Root Mean Squared Error
func calculateRMSE(formula *Node, examples []TrainingExample) float64 {
	sum := 0.0
	for _, ex := range examples {
		predicted := formula.Evaluate(getVariableMap(ex.Inputs))
		actual := ex.TargetReturn
		diff := predicted - actual
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(examples)))
}

// calculateSpearman calculates negative Spearman correlation (for minimization)
// Returns 1.0 - correlation, so lower is better
func calculateSpearman(formula *Node, examples []TrainingExample) float64 {
	if len(examples) < 2 {
		return 1.0 // Worst correlation
	}

	// Get predicted and actual values
	predicted := make([]float64, len(examples))
	actual := make([]float64, len(examples))

	for i, ex := range examples {
		predicted[i] = formula.Evaluate(getVariableMap(ex.Inputs))
		actual[i] = ex.TargetReturn
	}

	// Calculate Spearman correlation
	correlation := spearmanCorrelation(predicted, actual)

	// Return 1.0 - correlation (so lower is better, and we maximize correlation)
	return 1.0 - correlation
}

// spearmanCorrelation calculates Spearman rank correlation
func spearmanCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	// Rank the values
	rankX := rank(x)
	rankY := rank(y)

	// Calculate Pearson correlation of ranks
	return pearsonCorrelation(rankX, rankY)
}

// rank assigns ranks to values (handles ties by averaging)
func rank(values []float64) []float64 {
	ranks := make([]float64, len(values))

	// Create index-value pairs
	type pair struct {
		index int
		value float64
	}

	pairs := make([]pair, len(values))
	for i, v := range values {
		pairs[i] = pair{i, v}
	}

	// Sort by value
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].value > pairs[j].value {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// Assign ranks (handle ties)
	for i := 0; i < len(pairs); i++ {
		// Count ties
		tieCount := 1
		tieSum := float64(i + 1)
		for j := i + 1; j < len(pairs) && pairs[i].value == pairs[j].value; j++ {
			tieCount++
			tieSum += float64(j + 1)
		}

		// Average rank for ties
		avgRank := tieSum / float64(tieCount)
		for j := 0; j < tieCount; j++ {
			ranks[pairs[i+j].index] = avgRank
		}

		i += tieCount - 1
	}

	return ranks
}

// pearsonCorrelation calculates Pearson correlation coefficient
func pearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	// Calculate means
	meanX := 0.0
	meanY := 0.0
	for i := 0; i < len(x); i++ {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= float64(len(x))
	meanY /= float64(len(x))

	// Calculate covariance and variances
	cov := 0.0
	varX := 0.0
	varY := 0.0

	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		cov += dx * dy
		varX += dx * dx
		varY += dy * dy
	}

	if varX == 0 || varY == 0 {
		return 0.0
	}

	return cov / math.Sqrt(varX*varY)
}

// getVariableMap converts TrainingInputs to a variable map for formula evaluation
func getVariableMap(inputs TrainingInputs) map[string]float64 {
	return map[string]float64{
		"long_term":       inputs.LongTermScore,
		"fundamentals":    inputs.FundamentalsScore,
		"dividends":       inputs.DividendsScore,
		"opportunity":     inputs.OpportunityScore,
		"short_term":      inputs.ShortTermScore,
		"technicals":      inputs.TechnicalsScore,
		"opinion":         inputs.OpinionScore,
		"diversification": inputs.DiversificationScore,
		"total_score":     inputs.TotalScore,
		"cagr":            inputs.CAGR,
		"dividend_yield":  inputs.DividendYield,
		"volatility":      inputs.Volatility,
		"regime":          inputs.RegimeScore,
		"sharpe":          getFloatValue(inputs.SharpeRatio),
		"sortino":         getFloatValue(inputs.SortinoRatio),
		"rsi":             getFloatValue(inputs.RSI),
		"max_drawdown":    getFloatValue(inputs.MaxDrawdown),
	}
}

// getFloatValue safely gets float value from pointer
func getFloatValue(ptr *float64) float64 {
	if ptr == nil {
		return 0.0
	}
	return *ptr
}

// CalculateComplexity calculates the complexity of a formula (number of nodes)
func CalculateComplexity(formula *Node) int {
	if formula == nil {
		return 0
	}

	complexity := 1 // Count this node

	if formula.Left != nil {
		complexity += CalculateComplexity(formula.Left)
	}
	if formula.Right != nil {
		complexity += CalculateComplexity(formula.Right)
	}

	return complexity
}
