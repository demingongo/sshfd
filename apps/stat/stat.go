package stat

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/demingongo/sshfd/globals"
	"github.com/demingongo/sshfd/utils"

	"github.com/spf13/viper"
)

type DiskStat struct {
	Filesystem string
	Type       string
	Size       string
	Used       string
	Available  string
	UsePercent string
	MountedOn  string
}

func (m DiskStat) String() string {
	return m.MountedOn + " (" + m.Type + ") " + m.Used + "/" + m.Size + "  -  " + m.UsePercent
}

type MemStat struct {
	Type      string
	Total     string
	Used      string
	Free      string
	Shared    string
	Cache     string
	Available string
}

func (m MemStat) String() string {
	return m.Type + " " + m.Used + "/" + m.Total
}

var (
	subtle  = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	special = lipgloss.AdaptiveColor{Light: "230", Dark: "#010102"}

	// Spinner.
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	// Table cell.
	tableCellStyle      = lipgloss.NewStyle().Align(lipgloss.Center).PaddingLeft(1).PaddingRight(1)
	tableCellLeftStyle  = lipgloss.NewStyle().Align(lipgloss.Left).PaddingLeft(1).PaddingRight(1)
	tableCellRightStyle = lipgloss.NewStyle().Align(lipgloss.Right).PaddingLeft(1).PaddingRight(1)

	// Titles.

	titleStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(0, 1).
			Background(lipgloss.Color("7")).
			Foreground(special)

	subtitleStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle).
			Foreground(lipgloss.Color("6")).
			Bold(true)

	// Info block.

	infoStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)
)

func Run() {
	logger := globals.Logger

	if val, ok := utils.LoadHostConfig(viper.GetString("host")); ok && val.Hostname != "" {

		client, err := utils.DialSsh(val)
		if err != nil {
			logger.Fatalf("Unable to connect: %v", err)
		}
		defer client.Close()

		var disksStats []DiskStat
		var memStats []MemStat

		session, err := utils.CreateSession(client)
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer session.Close()

		if err := utils.RequestPty(session); err != nil {
			logger.Fatalf("Request for pseudo terminal failed: %v", err)
		}

		var b bytes.Buffer
		session.Stdout = &b // get output

		if err := session.Run("df -Th"); err != nil {
			logger.Error(b.String())
			logger.Fatalf("Failed to run: %v", err)
		}

		logger.Debugf("\n%v", b.String())

		// get the lines and remove the first line ([1:]) as it is the columns header
		dfLines := strings.Split(strings.ReplaceAll(b.String(), "\r\n", "\n"), "\n")[1:]

		for _, line := range dfLines {
			cols := filter(
				strings.Split(strings.Trim(line, ""), " "),
				isNotEmpty,
			)

			/*
			* 0 = Filesystem
			* 1 = Type
			* 2 = Size
			* 3 = Used
			* 4 = Available
			* 5 = Use%
			* 6 = Mounted on
			 */

			if len(cols) < 7 {
				continue
			}

			if cols[1] == "tmpfs" || cols[1] == "devtmpfs" || cols[1] == "efivarfs" {
				continue
			}

			disksStats = append(disksStats, DiskStat{
				Filesystem: cols[0],
				Type:       cols[1],
				Size:       cols[2],
				Used:       cols[3],
				Available:  cols[4],
				UsePercent: cols[5],
				MountedOn:  cols[6],
			})
		}

		logger.Debug(fmt.Sprintf("%v", disksStats))

		mSession, err := utils.CreateSession(client)
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer mSession.Close()

		if err := utils.RequestPty(mSession); err != nil {
			logger.Fatalf("Request for pseudo terminal failed: %v", err)
		}

		b.Reset() // empty buffer
		mSession.Stdout = &b

		if err := mSession.Run("free -mh"); err != nil {
			logger.Error(b.String())
			logger.Fatalf("Failed to run: %v", err)
		}

		logger.Debugf("\n%v", b.String())

		// get the lines and remove the first line ([1:]) as it is the columns header
		memLines := strings.Split(strings.ReplaceAll(b.String(), "\r\n", "\n"), "\n")[1:]

		for _, line := range memLines {
			cols := filter(
				strings.Split(strings.Trim(line, ""), " "),
				isNotEmpty,
			)

			if len(cols) < 4 {
				continue
			}

			if cols[1] == "total" {
				continue
			}

			/*
			* 0 = Type
			* 1 = Total
			* 2 = Used
			* 3 = Free
			* 4 = Shared
			* 5 = Cache
			* 6 = Available
			 */

			s := MemStat{
				Type:  strings.TrimSuffix(cols[0], ":"),
				Total: cols[1],
				Used:  cols[2],
				Free:  cols[3],
			}

			for i, metric := range cols[4:] {
				if i == 0 {
					s.Shared = metric
				} else if i == 1 {
					s.Cache = metric
				} else if i == 2 {
					s.Available = metric
				} else {
					break
				}
			}

			memStats = append(memStats, s)
		}

		logger.Debug(fmt.Sprintf("%v", memStats))

		cpuSession, err := utils.CreateSession(client)
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer cpuSession.Close()

		if err := utils.RequestPty(cpuSession); err != nil {
			logger.Fatalf("Request for pseudo terminal failed: %v", err)
		}

		b.Reset() // empty buffer
		cpuSession.Stdout = &b

		cpuPercent := float32(0)

		var cpuErr error

		_ = spinner.New().
			Type(spinner.MiniDot).
			Title(spinnerStyle.Render(" Please wait ...")).
			Style(spinnerStyle).
			Action(func() {
				if err := cpuSession.Run("grep --max-count=1 '^cpu.' /proc/stat && sleep 1 && grep --max-count=1 '^cpu.' /proc/stat"); err != nil {
					logger.Error(b.String())
					cpuErr = err
				}

				logger.Debugf("\n%v", b.String())

				// get cpu metrics
				cpuLines := strings.Split(strings.ReplaceAll(b.String(), "\r\n", "\n"), "\n")

				totalPrev := 0
				idlePrev := 0

				for _, line := range cpuLines {
					cpuMetrics := filter(
						strings.Split(strings.Trim(line, ""), " "),
						isNotEmpty,
					)

					if len(cpuMetrics) < 8 {
						break
					}

					cpuMetrics = cpuMetrics[1:]

					logger.Debugf("cpuMetrics %v", cpuMetrics)

					cpuUser, err := strconv.Atoi(cpuMetrics[0])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuNice, err := strconv.Atoi(cpuMetrics[1])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuSystem, err := strconv.Atoi(cpuMetrics[2])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuIdle, err := strconv.Atoi(cpuMetrics[3])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuIOwait, err := strconv.Atoi(cpuMetrics[4])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuIrq, err := strconv.Atoi(cpuMetrics[5])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuSoftirq, err := strconv.Atoi(cpuMetrics[6])
					if err != nil {
						logger.Error(err)
						break
					}
					cpuSteal, err := strconv.Atoi(cpuMetrics[7])
					if err != nil {
						logger.Error(err)
						break
					}

					cpuTotal := cpuUser + cpuNice + cpuSystem + cpuIdle + cpuIOwait + cpuIrq + cpuSoftirq + cpuSteal

					diffIdle := cpuIdle - idlePrev
					diffTotal := cpuTotal - totalPrev
					cpuPercent = (float32(1000*(diffTotal-diffIdle)) / float32(diffTotal+5)) / 10

					totalPrev = cpuTotal
					idlePrev = cpuIdle
				}
			}).
			Run()

		if cpuErr != nil {
			logger.Fatalf("Failed to run: %v", cpuErr)
		}

		logger.Debugf("%v", cpuPercent)

		// It should be the largest content (the title will be on multiple lines if larger).
		// Use it's width for later content.
		disksTable := table.New().Border(lipgloss.NormalBorder()).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch {
				case row != table.HeaderRow && col == 0:
					return tableCellLeftStyle
				case row != table.HeaderRow && col == 3:
					return tableCellRightStyle
				default:
					return tableCellStyle
				}
			}).
			Headers("", "Type", "Used", "%").
			Rows(ArrayMap(disksStats, func(v DiskStat) []string {
				return []string{v.MountedOn, v.Type, v.Used + "/" + v.Size, v.UsePercent}
			})...).Render()

		// the largest width
		largestWidth := lipgloss.Width(disksTable)

		content := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Width(largestWidth).Render(fmt.Sprintf("STAT %s", val.Host)),
			subtitleStyle.Width(largestWidth).Render("CPU "),
			strconv.FormatFloat(float64(cpuPercent), 'f', 2, 32)+"%",
			subtitleStyle.Width(largestWidth).Render("Memory"),
			table.New().Border(lipgloss.NormalBorder()).
				StyleFunc(func(row, col int) lipgloss.Style {
					switch {
					case row != table.HeaderRow && col == 0:
						return tableCellLeftStyle
					default:
						return tableCellRightStyle
					}
				}).
				Headers("Type", "Used", "Total").
				Rows(ArrayMap(memStats, func(v MemStat) []string {
					return []string{v.Type, v.Used, v.Total}
				})...).Render(),
			subtitleStyle.Width(largestWidth).Render("Disks"),
			disksTable,
		)

		result := infoStyle.
			PaddingLeft(2).
			PaddingRight(2).
			Render(content)

		fmt.Println(result)
	} else {
		logger.Fatal("No host")
	}
}

func isNotEmpty(s string) bool { return s != "" }

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func ArrayMap[TSource, TTarget any](source []TSource, trans func(TSource) TTarget) []TTarget {
	target := make([]TTarget, 0, len(source))
	for _, s := range source {
		target = append(target, trans(s))
	}
	return target
}
