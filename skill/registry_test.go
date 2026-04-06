package skill

import (
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()

	s := &Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Tags:        []string{"testing"},
	}

	r.Register(s)

	got := r.Get("test-skill")
	if got == nil {
		t.Fatal("Get returned nil for registered skill")
	}
	if got.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", got.Name, "test-skill")
	}
}

func TestRegistry_GetNonExistent(t *testing.T) {
	r := NewRegistry()

	got := r.Get("does-not-exist")
	if got != nil {
		t.Errorf("Get returned non-nil for unregistered skill: %v", got)
	}
}

func TestRegistry_RegisterNilSkill(t *testing.T) {
	r := NewRegistry()

	// nil 스킬 등록 시 패닉 없이 무시되어야 함
	r.Register(nil)

	if r.Count() != 0 {
		t.Errorf("Count = %d after registering nil, want 0", r.Count())
	}
}

func TestRegistry_RegisterEmptyNameSkill(t *testing.T) {
	r := NewRegistry()

	s := &Skill{Name: "", Description: "empty name"}
	r.Register(s)

	if r.Count() != 0 {
		t.Errorf("Count = %d after registering empty name, want 0", r.Count())
	}
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()

	skills := []*Skill{
		{Name: "skill-a", Description: "A"},
		{Name: "skill-b", Description: "B"},
		{Name: "skill-c", Description: "C"},
	}

	r.RegisterAll(skills)

	if r.Count() != 3 {
		t.Errorf("Count = %d, want 3", r.Count())
	}
}

func TestRegistry_CountAfterDuplicateRegister(t *testing.T) {
	r := NewRegistry()

	s := &Skill{Name: "dup-skill", Description: "Original"}
	r.Register(s)
	r.Register(s)

	// 같은 이름으로 덮어써도 Count는 1
	if r.Count() != 1 {
		t.Errorf("Count = %d after duplicate register, want 1", r.Count())
	}
}

func TestRegistry_GetByTag(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "skill-with-tag",
		Tags: []string{"frontend", "react"},
	})
	r.Register(&Skill{
		Name: "skill-without-tag",
		Tags: []string{"backend"},
	})

	results := r.GetByTag("frontend")
	if len(results) != 1 {
		t.Fatalf("GetByTag(frontend) returned %d results, want 1", len(results))
	}
	if results[0].Name != "skill-with-tag" {
		t.Errorf("GetByTag result Name = %q, want %q", results[0].Name, "skill-with-tag")
	}
}

func TestRegistry_GetByTagNoMatch(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "some-skill",
		Tags: []string{"backend"},
	})

	results := r.GetByTag("nonexistent")
	if len(results) != 0 {
		t.Errorf("GetByTag(nonexistent) returned %d results, want 0", len(results))
	}
}

func TestRegistry_GetByKeyword(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "auth-skill",
		Triggers: Triggers{
			Keywords: []string{"authentication", "JWT"},
		},
	})
	r.Register(&Skill{
		Name: "db-skill",
		Triggers: Triggers{
			Keywords: []string{"database", "SQL"},
		},
	})

	// 대소문자 무관 검색
	results := r.GetByKeyword("jwt")
	if len(results) != 1 {
		t.Fatalf("GetByKeyword(jwt) returned %d results, want 1", len(results))
	}
	if results[0].Name != "auth-skill" {
		t.Errorf("GetByKeyword result Name = %q, want %q", results[0].Name, "auth-skill")
	}
}

func TestRegistry_GetByKeywordCaseInsensitive(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "test-skill",
		Triggers: Triggers{
			Keywords: []string{"React"},
		},
	})

	// 소문자로 검색해도 찾아야 함
	results := r.GetByKeyword("react")
	if len(results) != 1 {
		t.Errorf("GetByKeyword(react) returned %d results, want 1", len(results))
	}
}

func TestRegistry_GetByPhase(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "plan-skill",
		Triggers: Triggers{
			Phases: []string{"plan"},
		},
	})
	r.Register(&Skill{
		Name: "run-skill",
		Triggers: Triggers{
			Phases: []string{"run"},
		},
	})

	results := r.GetByPhase("plan")
	if len(results) != 1 {
		t.Fatalf("GetByPhase(plan) returned %d results, want 1", len(results))
	}
	if results[0].Name != "plan-skill" {
		t.Errorf("GetByPhase result Name = %q, want %q", results[0].Name, "plan-skill")
	}
}

func TestRegistry_GetByAgent(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "backend-skill",
		Triggers: Triggers{
			Agents: []string{"backend", "fullstack"},
		},
	})

	results := r.GetByAgent("backend")
	if len(results) != 1 {
		t.Fatalf("GetByAgent(backend) returned %d, want 1", len(results))
	}
	if results[0].Name != "backend-skill" {
		t.Errorf("Name = %q, want %q", results[0].Name, "backend-skill")
	}
}

func TestRegistry_GetByLanguage(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "go-skill",
		Triggers: Triggers{
			Languages: []string{"Go", "Rust"},
		},
	})

	// 대소문자 무관
	results := r.GetByLanguage("go")
	if len(results) != 1 {
		t.Fatalf("GetByLanguage(go) returned %d, want 1", len(results))
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{
		Name: "to-be-cleared",
		Tags: []string{"temp"},
		Triggers: Triggers{
			Keywords: []string{"clear"},
		},
	})

	if r.Count() != 1 {
		t.Fatalf("Count before clear = %d, want 1", r.Count())
	}

	r.Clear()

	if r.Count() != 0 {
		t.Errorf("Count after clear = %d, want 0", r.Count())
	}
	if got := r.Get("to-be-cleared"); got != nil {
		t.Error("Get should return nil after Clear")
	}
	if tags := r.GetByTag("temp"); len(tags) != 0 {
		t.Error("GetByTag should return empty after Clear")
	}
}

func TestRegistry_AllSorted(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{Name: "charlie"})
	r.Register(&Skill{Name: "alpha"})
	r.Register(&Skill{Name: "bravo"})

	sorted := r.AllSorted()
	if len(sorted) != 3 {
		t.Fatalf("AllSorted returned %d skills, want 3", len(sorted))
	}
	if sorted[0].Name != "alpha" {
		t.Errorf("sorted[0] = %q, want alpha", sorted[0].Name)
	}
	if sorted[1].Name != "bravo" {
		t.Errorf("sorted[1] = %q, want bravo", sorted[1].Name)
	}
	if sorted[2].Name != "charlie" {
		t.Errorf("sorted[2] = %q, want charlie", sorted[2].Name)
	}
}

func TestRegistry_AllTags(t *testing.T) {
	r := NewRegistry()

	r.Register(&Skill{Name: "s1", Tags: []string{"b-tag", "a-tag"}})
	r.Register(&Skill{Name: "s2", Tags: []string{"c-tag", "a-tag"}})

	tags := r.AllTags()
	if len(tags) != 3 {
		t.Fatalf("AllTags returned %d tags, want 3", len(tags))
	}
	// 정렬 확인
	if tags[0] != "a-tag" {
		t.Errorf("tags[0] = %q, want a-tag", tags[0])
	}
}
