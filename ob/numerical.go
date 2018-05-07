package ob

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math"
	"strconv"

	"github.com/zhengxiaoyao0716/zmodule"

	"github.com/zhengxiaoyao0716/util/console"
	"github.com/zhengxiaoyao0716/util/cout"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/event"
)

var (
	// 首先，解决水陆比问题，在(0, 1)的随机范围内取水平面
	level = 0.7 // 海洋71%
	// 然后，防止碎片化，高地势聚合，高斯拟合
	gatherN = 7 // 七大洲
	// 考虑到面积多样化，各聚合力不同，对各大洲面积占比拟合
	// gatherFn = func(x float64) float64 { return 0.3*x*x - 0.5*x + 0.3 }
	gatherExpr = "0.3*x*x - 0.5*x + 0.3"
	// 考虑到陆地海拔分布，随机数 (level, 1) 的范围，映射到高度上（单位米）
	// 水平面0，总平均海拔850米，最高高度8844米，参照海陆起伏曲线进行幂函数拟合
	// hightFn = func(x float64) float64 { return 9000 * math.Pow(((x-level)/(1-level)), 3.5) }
	hightExpr = "9000 * math.Pow((x-level)/(1-level), 3.5)"
	// 类似的，考虑到海洋深度分布，随机数 (0, level) 的范围，映射到深度
	// 参照海陆起伏曲线，暂且忽略海沟，将平均深度按4000m算，仅拟合线性部分
	// depthFn    = func(x float64) float64 { return 8000 * (x - level) / (level) }
	depthExpr = "8000 * (x - level) / (level)"
	// 先将球面划分出1k个区块吧
	sampleN = 1000
)

var (
	gatherFn func(x float64) float64
	hightFn  func(x float64) float64
	depthFn  func(x float64) float64
)

func initNumerical() {
	gatherFn = exprFn("gatherExpr", gatherExpr)
	hightFn = exprFn("hightExpr", hightExpr)
	depthFn = exprFn("depthExpr", depthExpr)
	// gatherFn = func(x float64) float64 { return 0.3*x*x - 0.5*x + 0.3 }
	// hightFn = func(x float64) float64 { return 9000 * math.Pow(((x-level)/(1-level)), 3.5) }
	// depthFn = func(x float64) float64 { return 8000 * (x - level) / (level) }
}

func exprFn(name, expr string) func(x float64) float64 {
	astExpr, err := parser.ParseExpr(expr)
	if err != nil {
		log.Fatalln(err)
	}
	// ast.Print(nil, astExpr)
	return func(x float64) float64 { return evalExpr(expr, astExpr, x) }
}

func evalExpr(raw string, expr ast.Expr, x float64) float64 {
	failed := func(p token.Pos, err interface{}) {
		console.Log(raw[0:p-1] + cout.Err("%c", raw[p-1]) + raw[p:])
		log.Fatalln(err)
	}
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		ex := evalExpr(raw, expr.X, x)
		ey := evalExpr(raw, expr.Y, x)
		var r float64
		switch expr.Op {
		case token.ADD:
			r = ex + ey
		case token.SUB:
			r = ex - ey
		case token.MUL:
			r = ex * ey
		case token.QUO:
			r = ex / ey
		}
		return r
	case *ast.BasicLit:
		var (
			r   float64
			err error
		)
		switch expr.Kind {
		case token.INT:
			var i int
			i, err = strconv.Atoi(expr.Value)
			r = float64(i)
		case token.FLOAT:
			r, err = strconv.ParseFloat(expr.Value, 64)
		default:
			failed(expr.Pos(), "nonsupport basic kind: "+expr.Kind.String())
		}
		if err != nil {
			failed(expr.Pos(), err)
		}
		return r
	case *ast.Ident:
		switch expr.Name {
		case "x":
			return x
		case "level":
			return level
		case "gatherN":
			return float64(gatherN)
		default:
			failed(expr.Pos(), "unknown ident name: "+expr.Name)
		}
	case *ast.CallExpr:
		switch fun := expr.Fun.(type) {
		case *ast.SelectorExpr:
			switch X := fun.X.(type) {
			case *ast.Ident:
				switch X.Name {
				case "math":
					switch fun.Sel.Name {
					case "Pow":
						return math.Pow(evalExpr(raw, expr.Args[0], x), evalExpr(raw, expr.Args[1], x))
					default:
						failed(fun.Sel.Pos(), "nonsupport method: "+fun.Sel.Name)
					}
				default:
					failed(X.Pos(), "nonsupport module: "+X.Name)
				}
			default:
				failed(X.Pos(), "invalid expression type: "+cout.Err("%T", expr))
			}
		default:
			failed(fun.Pos(), "invalid expression type: "+cout.Err("%T", expr))
		}
	case *ast.ParenExpr:
		return evalExpr(raw, expr.X, x)
	default:
		failed(expr.Pos(), "invalid expression type: "+cout.Err("%T", expr))
	}

	return 0
}

func init() {
	zmodule.Args["level"] = zmodule.Argument{
		Default: level,
		Usage:   "Average level for earth.",
	}
	zmodule.Args["gatherN"] = zmodule.Argument{
		Default: gatherN,
		Usage:   "Number of expect gather center.",
	}
	zmodule.Args["gatherExpr"] = zmodule.Argument{
		Default: gatherExpr,
		Usage:   "Polyfit expression for `gather` function.",
	}
	zmodule.Args["hightExpr"] = zmodule.Argument{
		Default: hightExpr,
		Usage:   "Polyfit expression for `hight` function.",
	}
	zmodule.Args["depthExpr"] = zmodule.Argument{
		Default: depthExpr,
		Usage:   "Polyfit expression for `depth` function.",
	}
	zmodule.Args["sampleN"] = zmodule.Argument{
		Default: sampleN,
		Usage:   "Number of samples point.",
	}
	event.OnInit(func(event.Event) error {
		event.On(event.KeyStart, func(event.Event) error {
			level = config.GetNum("level")
			gatherN = config.GetInt("gatherN")
			gatherExpr = config.GetString("gatherExpr")
			hightExpr = config.GetString("hightExpr")
			depthExpr = config.GetString("depthExpr")
			sampleN = config.GetInt("sampleN")
			initNumerical()
			return nil
		})
		return nil
	})
}
