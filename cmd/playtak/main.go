package main

import (
	"fmt"
	"bufio"
	"flag"
	//"io/ioutil"
	"log"
	//"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"../../ai"
	"../../ai/mcts"
	"../../cli"
	//"../../ptn"
	"../../tak"
)

var (
	white = flag.String("white", "human", "white player")
	black = flag.String("black", "nohat", "white player")
	size  = flag.Int("size", 5, "game size")
	debug = flag.Int("debug", 0, "debug level")
	limit = flag.Duration("limit", time.Minute, "ai time limit")
	out   = flag.String("out", "", "write ptn to file")
	repeat = flag.Int("repeat", 1000, "number of games")
	silent = flag.Bool("silent", false, "print nothing")
)

type aiWrapper struct {
	p ai.TakPlayer
}

func (a *aiWrapper) GetMove(p *tak.Position) tak.Move {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(*limit))
	defer cancel()
	return a.p.GetMove(ctx, p)
}

func parsePlayer(in *bufio.Reader, s string) cli.Player {
	if s == "human" {
		return cli.NewCLIPlayer(os.Stdout, in)
	}
	if s == "rand" {
		return &aiWrapper{ai.NewRandom(0)}
	}
	if strings.HasPrefix(s, "rand") {
		var seed int64
		if len(s) > len("rand") {
			i, err := strconv.Atoi(s[len("rand:"):])
			if err != nil {
				log.Fatal(err)
			}
			seed = int64(i)
		}
		return &aiWrapper{ai.NewRandom(seed)}
	}
	if s == "nohat" {
		p := ai.NewMinimax(ai.MinimaxConfig{
			Size:  *size,
			Debug: *debug,
			Depth: 3,
			Evaluate: ai.MakeNohat(*size, nil),
			NoTable: true,
		})
		return &aiWrapper{p}
	}
	if strings.HasPrefix(s, "minimax") {
		var depth = 3
		if len(s) > len("minimax") {
			i, err := strconv.Atoi(s[len("minimax:"):])
			if err != nil {
				log.Fatal(err)
			}
			depth = i
		}
		p := ai.NewMinimax(ai.MinimaxConfig{
			Size:  *size,
			Depth: depth,
			Debug: *debug,
			NoTable: true,
			
			//Evaluate: ai.MakeEvaluator(5, &ai.Weights{TopFlat: 100}),
		})
		return &aiWrapper{p}
	}
	if strings.HasPrefix(s, "mcts") {
		var limit = 30 * time.Second
		if len(s) > len("mcts") {
			var err error
			limit, err = time.ParseDuration(s[len("mcts:"):])
			if err != nil {
				log.Fatal(err)
			}
		}
		p := mcts.NewMonteCarlo(mcts.MCTSConfig{
			Limit: limit,
			Debug: *debug,
			Size:  *size,
		})
		return &aiWrapper{p}
	}
	log.Fatalf("unparseable player: %s", s)
	return nil
}

func newNohat(size int, w []float64) cli.Player {
	p := ai.NewMinimax(ai.MinimaxConfig{
		Size:  size,
		Debug: 0,
		Depth: 1,
		Evaluate: ai.MakeNohat(size, w),
		NoTable: true,
	})
	return &aiWrapper{p}
}
/*
//Adjust weights
func main(){
	weightin := []float64{-7.882857408056033, -18.422073883974566, -36.0493127331049, -79.56604291873225, -118.62470509229402, -210.29295370964735, -517.3034545361437, -100000, -8.152244675890454, 30.862485736157335, 53.218528069373136, 64.62304527497034, 168.4353775751191, 277.202690219266, 10000, 20000, 758.9873559223424, 583.7652877392484, 418.738596952849, 181.21979073578143, 0, 23.88262609962125, 95.00995285905964, 226.4340785427944, 790.8549324342757, -9.185201207821727, -31.290273494645234, 0.0028395103346359513, 5.204139819874648, 73.41388857816798}
	flag.Parse()
	//in := bufio.NewReader(os.Stdin)
	baseWeight := make([]float64, 30)
	modWeight := make([]float64, 30)
	//var baseWeight []float64
	//var modWeight []float64
	copy(baseWeight,weightin)
	copy(modWeight,weightin)
	gameresult := make(chan int) //1: p1, 2: draw, 3: p2
	throttle := make(chan int, 8)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//f, _ := os.Create("weights.txt")
	f, _ := os.OpenFile("weights.txt", os.O_APPEND|os.O_WRONLY, 0600)
	playgame := func (w1 []float64, w2 []float64, swap bool) {
		if swap {
			w3 := w1
			w1 = w2
			w2 = w3
		}
		st := &cli.CLI{
			Config: tak.Config{Size: *size},
			Out:    ioutil.Discard, //os.Stdout, //
			White:  newNohat(*size, w1),
			Black:  newNohat(*size, w2),
			Silent: true,
		}
		final := st.Play()
		if final.WinDetails().Winner == tak.White {
			if swap {
				gameresult <- 3
			} else {
				gameresult <- 1
			}
		} else if final.WinDetails().Winner == tak.Black {
			if swap {
				gameresult <- 1
			} else {
				gameresult <- 3
			}
		} else {
			gameresult <- 2
		}
		<- throttle
	}
	play100 := func () {
		for a:=0; a<100; a++ {
			//fmt.Print("1")
			throttle <- 1
			//fmt.Print("2")
			go playgame(baseWeight, modWeight, a&1==1)
		}
	}
	run := 0
	changed := 0
	for {
		for a:=0; a<30; a++ {
			if r.Float64()>.5 {
				modWeight[a]=baseWeight[a]+(r.Float64()-.5)*.1*ai.WeightNohatAdjust[a]
			} else {
				modWeight[a]=baseWeight[a]
			}
		}
		score := int(0)
		for score<30 && score>-30 {
			go play100()
			for a:=0; a<100; a++ {
				score += <- gameresult - 2
			}
			fmt.Printf("%d\n",score)
		}
		run++
		if score>=30 {
			changed++
			for a:=0; a<30; a++ {
				baseWeight[a]=baseWeight[a]*.5+modWeight[a]*.5
			}
			weightext := fmt.Sprintf("%#v\n\n",baseWeight)
			f.WriteString(weightext)
			weightext = fmt.Sprintf("%d / %d\n\n",changed,run)
			f.WriteString(weightext)
		}
		fmt.Printf("%#v\n\n",baseWeight)
	}
}
*/
/*
// Test play
func main(){
	weightin := []float64{0, -20.80992860246264, -39.53973239686991, -81.50246863507586, -155.93079444027765, -289.23796777428583, -614.5346962394407, -100000, 0, 32.30082505350642, 45.12214408862102, 65.61008053136833, 156.60239035239982, 413.263468400728, 10000, 20000, 589.4603825920285, 506.1123687112676, 344.8402195215042, 194.76452676703917, 0, 19.077561948722284, 98.2862020188589, 222.64556765567144, 612.5365342926455, -9.800679880630113, -36.74424919228099, 0.002582788992910993, 5.446911960287178, 72.91389999925025}
	flag.Parse()
	limit := *repeat
	in := bufio.NewReader(os.Stdin)
	baseWeight := make([]float64, 30)
	modWeight := make([]float64, 30)
	//var baseWeight []float64
	//var modWeight []float64
	copy(baseWeight,weightin)
	copy(modWeight,weightin)
	gameresult := make(chan int) //1: p1, 2: draw, 3: p2
	throttle := make(chan int, 8)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//f, _ := os.Create("weights.txt")
	playgame := func (w1 []float64, w2 []float64, swap bool) {
		if swap {
			w3 := w1
			w1 = w2
			w2 = w3
		}
		st := &cli.CLI{
			Config: tak.Config{Size: *size},
			Out:    ioutil.Discard, //os.Stdout, //
			White:  parsePlayer(in, *white),
			Black:  parsePlayer(in, *black),
			Silent: true,
		}
		final := st.Play()
		if final.WinDetails().Winner == tak.White {
			if swap {
				gameresult <- 3
			} else {
				gameresult <- 1
			}
		} else if final.WinDetails().Winner == tak.Black {
			if swap {
				gameresult <- 1
			} else {
				gameresult <- 3
			}
		} else {
			gameresult <- 2
		}
		<- throttle
	}
	play100 := func () {
		for a:=0; a<limit; a++ {
			//fmt.Print("1")
			throttle <- 1
			//fmt.Print("2")
			go playgame(baseWeight, modWeight, false)
		}
	}
	run := 0
	changed := 0
	for {
		for a:=0; a<30; a++ {
			if r.Float64()>.5 {
				modWeight[a]=baseWeight[a]+(r.Float64()-.5)*.04*ai.WeightNohatAdjust[a]
			} else {
				modWeight[a]=baseWeight[a]
			}
		}
		score := int(0)
		go play100()
		for a:=0; a<limit; a++ {
			score += <- gameresult - 2
		}
		fmt.Printf("%d\n",score)
		run++
		if score>=40 {
			changed++
			for a:=0; a<30; a++ {
				baseWeight[a]=baseWeight[a]*.5+modWeight[a]*.5
			}
			//weightext := fmt.Sprintf("%#v\n\n",baseWeight)
			//f.WriteString(weightext)
			//weightext = fmt.Sprintf("%d / %d\n\n",changed,run)
			//f.WriteString(weightext)
		}
		fmt.Printf("%#v\n\n",baseWeight)
	}
}
*/

func main() {
	flag.Parse()
	in := bufio.NewReader(os.Stdin)
	limit := *repeat
	result := ""
	for b:=0; b<5; b++ {
		//ai.WeightNohat[6] = -480 - 80 * float64(b)
		winsA := 0
		winsB := 0
		for a:=0; a<limit; a++ {
			st := &cli.CLI{
				Config: tak.Config{Size: *size},
				Out:    os.Stdout, //ioutil.Discard, //
				White:  parsePlayer(in, *white),
				Black:  parsePlayer(in, *black),
				Silent: *silent,
			}
			final := st.Play()
			/*if *out != "" {
				p := &ptn.PTN{}
				p.Tags = []ptn.Tag{
					{Name: "Size", Value: strconv.Itoa(*size)},
					{Name: "Player1", Value: *white},
					{Name: "Player2", Value: *white},
				}
				p.AddMoves(st.Moves())
				ioutil.WriteFile(*out, []byte(p.Render()), 0644)
			}*/
			fmt.Printf("%d.%d\n",b,a)
			fmt.Printf("%+v\n\n", final.WinDetails().Winner)
			fmt.Printf("%+v\n\n", final.GetHash())
			if final.WinDetails().Winner == tak.White {
				winsA++
			}
			if final.WinDetails().Winner == tak.Black {
				winsB++
			}
		}
		fmt.Printf("%d - %d\n",winsA,winsB)
		result += fmt.Sprintf("%d",winsA) + " - " + fmt.Sprintf("%d",winsB) + "\n"
	}
	fmt.Printf("%#v\n\n",ai.WeightNohat)
	fmt.Print(result)
}
