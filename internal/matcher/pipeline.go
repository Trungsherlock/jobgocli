package matcher

import (
	"github.com/Trungsherlock/jobgo/internal/database"
	"github.com/spf13/viper"
)

type ScoringMode string

const (
	ModeKeyword	ScoringMode = "keyword"
	ModeLLM		ScoringMode = "llm"
	ModeHybrid	ScoringMode = "hybrid"
)

type Pipeline struct {
	keyword		*SkillScorer
	llm			*LLMSkillScorer
	mode 		ScoringMode
	threshold	float64
}

func NewPipeline() *Pipeline {
	mode := ScoringMode(viper.GetString("matcher.type"))
	if mode == "" {
		mode = ModeKeyword
	}
	threshold := viper.GetFloat64("matcher.llm_threshold")
	if threshold == 0 {
		threshold = 30
	}
	apiKey := viper.GetString("anthropic_api_key")

	p := &Pipeline{
		keyword:	NewSkillScorer(),
		mode: 		mode,
		threshold:	threshold,
	}
	if (mode == ModeLLM || mode == ModeHybrid) && apiKey != "" {
		p.llm = NewLLMSkillScorer(apiKey)
	}
	return p
}

func (p *Pipeline) Score(job database.Job, profile database.Profile) SkillScoreResult {
	switch p.mode {
	case ModeLLM:
		if p.llm != nil {
			if result, err := p.llm.Score(job, profile); err == nil {
				return result
			}
		}
		return p.keyword.Score(job, profile)
	case ModeHybrid:
		keywordResult := p.keyword.Score(job, profile)
		if p.llm != nil && keywordResult.Score >= p.threshold {
			if result, err := p.llm.Score(job, profile); err == nil {
				return result
			}
		}
		return keywordResult
		
	default:
		return p.keyword.Score(job, profile)
	}
}