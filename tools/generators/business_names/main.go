package main

// This file is used to generate the business_names.txt output file
import (
	"bufio"
	"fmt"
	"os"
)

var industries = []string{
	"Tech", "Global", "Systems", "Solutions", "Industries", "Enterprises", "Digital",
	"Innovation", "Dynamics", "Networks", "Technologies", "Software", "Analytics",
	"Services", "Group", "Partners", "International", "Ventures", "Capital", "Labs",
}

var prefixes = []string{
	"Advanced", "Allied", "Alpine", "Atlas", "Apex", "Blue", "Bright", "Clear",
	"Core", "Crown", "Crystal", "Delta", "Digital", "Elite", "First", "Future",
	"Global", "Golden", "Green", "Horizon", "Imperial", "Infinite", "Innovative",
	"Integrated", "Key", "Logic", "Matrix", "Meta", "Modern", "Neo", "New",
	"Next", "Nova", "Omega", "Optimal", "Peak", "Prime", "Quantum", "Rapid",
	"Royal", "Smart", "Solar", "Star", "Strategic", "Summit", "Supreme", "Sync",
	"Unity", "Universal", "Vector", "Vertex", "Vision", "Vital", "Wave", "World",
}

var baseWords = []string{
	"Action", "Advance", "Alpha", "Asset", "Basis", "Bridge", "Byte", "Cloud",
	"Connect", "Core", "Craft", "Cross", "Cyber", "Data", "Direct", "Edge",
	"Energy", "Flow", "Forge", "Form", "Fusion", "Grid", "Harbor", "Helix",
	"Hub", "Info", "Inter", "Link", "Logic", "Loop", "Mind", "Net", "Node",
	"Path", "Peak", "Plus", "Point", "Pulse", "Quest", "Reach", "Realm",
	"Scope", "Shift", "Spark", "Sphere", "Spire", "Stream", "Sync", "Sys",
	"Task", "Team", "Tech", "Tel", "Track", "Trade", "Trust", "Unity",
	"Urban", "Value", "Vector", "Vent", "Verge", "Verse", "Vibe", "View",
	"Vision", "Wave", "Way", "Web", "Wire", "Wise", "Work", "Zone",
	"Algo", "Array", "Audit", "Base", "Binary", "Block", "Cache", "Chain",
	"Code", "Crypt", "Dash", "Delta", "Dev", "Dock", "Drive", "Echo",
	"File", "Filter", "Flash", "Frame", "Gate", "Graph", "Hash", "Host",
	"Input", "Ion", "Jump", "Key", "Lab", "Lambda", "Layer", "Line",
	"Map", "Matrix", "Memory", "Mesh", "Meta", "Micro", "Monitor", "Neural",
	"Nexus", "Omega", "Orbit", "Parse", "Port", "Proxy", "Quantum", "Query",
	"Radio", "Ram", "Route", "Scale", "Script", "Server", "Signal", "Stack",
	"Terra", "Thread", "Token", "Trace", "Trunk", "Unit", "Upload", "Vertex",
	"Agile", "Axis", "Bank", "Board", "Bond", "Branch", "Brand", "Budget",
	"Cap", "Capital", "Central", "Channel", "Chart", "Chief", "Circle", "Claim",
	"Command", "Control", "Credit", "Crown", "Cube", "Cycle", "Domain", "Draft",
	"Eagle", "Engine", "Factor", "Field", "Fleet", "Force", "Forum", "Fund",
	"Garden", "Giant", "Global", "Growth", "Guard", "Guide", "Index", "Insight",
	"Iron", "Lead", "League", "Level", "Light", "Lion", "Market", "Medal",
	"Merit", "Method", "Model", "Motion", "Nova", "Ocean", "Office", "Order",
	"Panel", "Pearl", "Phase", "Plan", "Power", "Prime", "Profit", "Rank",
	"Ai", "Api", "App", "Arc", "Beam", "Bit", "Brain", "Bridge",
	"Cast", "Chain", "Click", "Cloud", "Core", "Cosmos", "Dart", "Data",
	"Edge", "Flex", "Flow", "Flux", "Gene", "Geo", "Giga", "Grid",
	"Hub", "Hyper", "Impact", "Intel", "Jet", "Kit", "Lens", "Link",
	"Max", "Mine", "Nano", "Net", "Neural", "Next", "Node", "Norm",
	"Optic", "Pixel", "Pod", "Pulse", "Ray", "Sage", "Scan", "Sense",
	"Smart", "Sol", "Swift", "Synth", "Titan", "Ultra", "Volt", "Zone",
}

var suffixes = []string{
	"Corp", "Inc", "LLC", "Ltd", "Group", "Holdings",
}

func main() {
	args := os.Args
	if len(args) < 2 {
		panic("must provide necessary args")
	}

	outputFile := args[1]

	if outputFile == "" {
		panic("output file is required")
	}

	// Call your generation function
	err := generateBusinessNames(industries, prefixes, baseWords, suffixes, outputFile)
	if err != nil {
		panic(err)
	}
}

func generateBusinessNames(industries, prefixes, baseWords, suffixes []string, fileName string) error {
	// create file to write to
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// create writer
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// helper func
	addName := func(name string) error {
		if name != "" {
			_, err := writer.WriteString(name + "\n")
			if err != nil {
				return fmt.Errorf("error writing to file: %v", err)
			}
		}
		return nil
	}

	// 1-word names
	for _, word := range baseWords {
		err := addName(word)
		if err != nil {
			return err
		}
	}

	// 2-word names
	// word + suffix
	for _, word := range industries {
		for _, suffix := range suffixes {
			err := addName(word + " " + suffix)
			if err != nil {
				return err
			}
		}
	}
	for _, word := range prefixes {
		for _, suffix := range suffixes {
			err := addName(word + " " + suffix)
			if err != nil {
				return err
			}
		}
	}

	for _, word := range baseWords {
		for _, suffix := range suffixes {
			err := addName(word + " " + suffix)
			if err != nil {
				return err
			}
		}
	}

	// word + word
	for _, word1 := range industries {
		for _, word2 := range prefixes {
			err := addName(word1 + " " + word2)
			if err != nil {
				return err
			}
		}
		for _, word2 := range baseWords {
			err := addName(word1 + " " + word2)
			if err != nil {
				return err
			}
		}
		for _, word2 := range industries {
			if word1 != word2 {
				err := addName(word1 + " " + word2)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, word1 := range prefixes {
		for _, word2 := range industries {
			err := addName(word1 + " " + word2)
			if err != nil {
				return err
			}
		}
		for _, word2 := range baseWords {
			err := addName(word1 + " " + word2)
			if err != nil {
				return err
			}
		}
		for _, word2 := range prefixes {
			if word1 != word2 {
				err := addName(word1 + " " + word2)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, word1 := range baseWords {
		for _, word2 := range industries {
			err := addName(word1 + " " + word2)
			if err != nil {
				return err
			}
		}
		for _, word2 := range prefixes {
			err := addName(word1 + " " + word2)
			if err != nil {
				return err
			}
		}
		for _, word2 := range baseWords {
			if word1 != word2 {
				err := addName(word1 + " " + word2)
				if err != nil {
					return err
				}
			}
		}
	}

	// 3-word names
	// word + word + suffix
	for _, word1 := range industries {
		for _, word2 := range prefixes {
			for _, suffix := range suffixes {
				err := addName(word1 + " " + word2 + " " + suffix)
				if err != nil {
					return err
				}
			}
		}
		for _, word2 := range baseWords {
			for _, suffix := range suffixes {
				err := addName(word1 + " " + word2 + " " + suffix)
				if err != nil {
					return err
				}
			}
		}
		for _, word2 := range industries {
			if word1 != word2 {
				for _, suffix := range suffixes {
					err := addName(word1 + " " + word2 + " " + suffix)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	for _, word1 := range prefixes {
		for _, word2 := range industries {
			for _, suffix := range suffixes {
				err := addName(word1 + " " + word2 + " " + suffix)
				if err != nil {
					return err
				}
			}
		}
		for _, word2 := range baseWords {
			for _, suffix := range suffixes {
				err := addName(word1 + " " + word2 + " " + suffix)
				if err != nil {
					return err
				}
			}
		}
		for _, word2 := range prefixes {
			if word1 != word2 {
				for _, suffix := range suffixes {
					err := addName(word1 + " " + word2 + " " + suffix)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	for _, word1 := range baseWords {
		for _, word2 := range industries {
			for _, suffix := range suffixes {
				err := addName(word1 + " " + word2 + " " + suffix)
				if err != nil {
					return err
				}
			}
		}
		for _, word2 := range prefixes {
			for _, suffix := range suffixes {
				err := addName(word1 + " " + word2 + " " + suffix)
				if err != nil {
					return err
				}
			}
		}
		for _, word2 := range baseWords {
			if word1 != word2 {
				for _, suffix := range suffixes {
					err := addName(word1 + " " + word2 + " " + suffix)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// word + word + word
	for _, word1 := range industries {
		for _, word2 := range prefixes {
			for _, word3 := range baseWords {
				err := addName(word1 + " " + word2 + " " + word3)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, word1 := range prefixes {
		for _, word2 := range industries {
			for _, word3 := range baseWords {
				err := addName(word1 + " " + word2 + " " + word3)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, word1 := range baseWords {
		for _, word2 := range industries {
			for _, word3 := range prefixes {
				err := addName(word1 + " " + word2 + " " + word3)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
