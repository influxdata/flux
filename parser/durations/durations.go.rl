package durations

import (
    "github.com/influxdata/flux/ast"
    "strconv"
    "fmt"
)

%%{
machine durations;

alphtype uint8;

action mark {
	m.pb = m.p
}

action rank_year {
    m.durations = append(m.durations, getDuration(m.text(), "y"))
    m.durationrank = 10
}

action rank_month {
    m.durations = append(m.durations, getDuration(m.text(), "mo"))
    m.durationrank = 9
}

action rank_week {
    m.durations = append(m.durations, getDuration(m.text(), "w"))
    m.durationrank = 8
}

action rank_day {
    m.durations = append(m.durations, getDuration(m.text(), "d"))    
    m.durationrank = 7
}

action rank_hour {
    m.durations = append(m.durations, getDuration(m.text(), "h"))    
    m.durationrank = 6
}

action rank_min {
    m.durations = append(m.durations, getDuration(m.text(), "m"))    
    m.durationrank = 5
}

action rank_sec {
    m.durations = append(m.durations, getDuration(m.text(), "s"))    
    m.durationrank = 4
}

action rank_millisec {
    m.durations = append(m.durations, getDuration(m.text(), "ms"))    
    m.durationrank = 3
}

action rank_us {
    m.durations = append(m.durations, getDuration(m.text(), "us"))    
    m.durationrank = 2
}

action rank_microsec {
    m.durations = append(m.durations, getDuration(m.text(), "μs"))    
    m.durationrank = 2
}

action rank_ns {
    m.durations = append(m.durations, getDuration(m.text(), "ns"))    
    m.durationrank = 1
}

action exit_durationsliteral {
    // re-init (e.g, reset) durations rank
    m.initDurationsRank()
    // populate expression
    m.expression = &ast.DurationLiteral{
        Values: m.durations,
    }
    // empty durations slice
    m.durations = nil
}

action exit_program {
    m.root = &ast.Program{
        Body: append([]ast.Statement{}, &ast.ExpressionStatement{
            Expression: m.expression,
        }),
        // todo > no base node for now => BaseNode: base(m.text(), m.curline, m.col()),
    }
}

non0digit = '1'..'9';

action generic_error {
    m.err = fmt.Errorf("unable to match [col %d]", m.p)
	fhold;
    fgoto fail;
}

duration = 
    start: (
        (non0digit . digit*) >mark -> units
    ),
    units: (
        ('y' when { m.durationrank > 10 } %rank_year) -> again |
        ('y' when { m.durationrank > 10 } %rank_year) -> final |

        ('mo' when { m.durationrank > 9 } @(mpriority, 2) %rank_month) -> again |
        ('mo' when { m.durationrank > 9 } %(mpriority, 2) %rank_month) -> final |

        ('w' when { m.durationrank > 8 } %rank_week) -> again |
        ('w' when { m.durationrank > 8 } %rank_week) -> final |

        ('d' when { m.durationrank > 7 } %rank_day) -> again |
        ('d' when { m.durationrank > 7 } %rank_day) -> final |

        ('h' when { m.durationrank > 6 } %rank_hour) -> again |
        ('h' when { m.durationrank > 6 } %rank_hour) -> final |

        ('m' when { m.durationrank > 5 } @(mpriority, 1) %rank_min) -> again |
        ('m' when { m.durationrank > 5 } %(mpriority, 1) %rank_min) -> final |

        ('s' when { m.durationrank > 4 } %rank_sec) -> again |
        ('s' when { m.durationrank > 4 } %rank_sec) -> final |

        ('ms' when { m.durationrank > 3 } @(mpriority, 2) %rank_millisec) -> again |
        ('ms' when { m.durationrank > 3 } %(mpriority, 2) %rank_millisec) -> final |

        ('us' when { m.durationrank > 2 } %rank_us) -> again |  
        ('us' when { m.durationrank > 2 } %rank_us) -> final |

        ((0xCE . 0xBC . 's') when { m.durationrank > 2 } %rank_microsec) -> again |  
        ((0xCE . 0xBC . 's') when { m.durationrank > 2 } %rank_microsec) -> final |
        
        ('ns' when { m.durationrank > 1 } %rank_ns) -> final
    ),
    again: (
        (non0digit . digit*) >mark -> units
    );

durationliteral = duration $err(generic_error) %eof(exit_durationsliteral);

main := durationliteral %exit_program;

fail := (any - space)* @err{ fgoto main; };

}%%

%% write data noerror noprefix;

type machine struct {
	data         []byte
	cs           int
	p, pe, eof   int
	pb           int
    curline      int
	err          error

    root *ast.Program
	expression   ast.Expression
	durations 	 []ast.Duration

	durationrank int
}

func NewMachine() *machine {
	m := &machine{}

	%% access m.;
	%% variable p m.p;
	%% variable pe m.pe;
	%% variable eof m.eof;
	%% variable data m.data;

	return m
}

// Err returns the error that occurred on the last call to Parse.
//
// If the result is nil, then the line was parsed successfully.
func (m *machine) Err() error {
	return m.err
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}

func (m *machine) initDurationsRank() {
    m.durationrank = 11
}

func getDuration(bytes []byte, unit string) ast.Duration {
	v1 := bytes[:len(bytes)-len(unit)]
	v2, _ := strconv.Atoi(string(v1))
	if unit == "μs" {
		unit = "us"
	}
	return ast.Duration{
		Magnitude: int64(v2),
		Unit:      unit,
	}
}

func (m *machine) Parse(input []byte) *ast.Program {
	m.data = input
	m.p = 0
	m.pb = 0
	m.pe = len(input)
	m.eof = len(input)
	m.err = nil
	m.root = nil
	m.initDurationsRank()

    %% write init;
    %% write exec;

	if m.cs < first_final  {
		return nil
	}

	return m.root
}