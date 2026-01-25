# JikiME-ADK Marketing Skills

마케팅 전문가 수준의 스킬 모음으로, 제품 런칭부터 전환 최적화까지 전체 마케팅 사이클을 지원합니다.

> **Attribution**: Enhanced from [marketingskills](https://github.com/coreyhaines/marketingskills) by Corey Haines (MIT License)

## 개요

JikiME-ADK 마케팅 스킬은 10개의 전문 영역을 커버합니다:

```
┌─────────────────────────────────────────────────────────────────┐
│                    MARKETING SKILL ECOSYSTEM                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  Launch  │───▶│   CRO    │───▶│ Analytics│───▶│ A/B Test │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│       │              │               │               │          │
│       ▼              ▼               ▼               ▼          │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  Email   │    │Copywrite │    │Onboarding│    │Psychology│  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│       │              │               │               │          │
│       └──────────────┴───────┬───────┴───────────────┘          │
│                              ▼                                   │
│                    ┌──────────────────┐                         │
│                    │  SEO  │ Pricing  │                         │
│                    └──────────────────┘                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 스킬 목록

| 스킬 | 설명 | 주요 활용 |
|------|------|----------|
| [page-cro](#1-page-cro) | 랜딩페이지 전환율 최적화 | 랜딩페이지 분석, CRO 전략 |
| [copywriting](#2-copywriting) | 전환 중심 카피라이팅 | 헤드라인, CTA, 세일즈 카피 |
| [pricing](#3-pricing) | 가격 전략 및 리서치 | Van Westendorp, MaxDiff |
| [psychology](#4-psychology) | 마케팅 심리 모델 | 70+ 멘탈 모델 |
| [seo](#5-seo) | SEO 감사 및 최적화 | 기술 SEO, 콘텐츠 SEO |
| [ab-test](#6-ab-test) | A/B 테스트 설계 | 가설 수립, 샘플 사이즈 |
| [email](#7-email) | 이메일 시퀀스 | 드립 캠페인, 라이프사이클 |
| [launch](#8-launch) | 제품 런칭 전략 | 5단계 런칭, ORB 프레임워크 |
| [analytics](#9-analytics) | 분석 추적 구현 | GA4, GTM, UTM |
| [onboarding](#10-onboarding) | 사용자 온보딩 | 활성화, Time-to-Value |

---

## 1. Page CRO

**스킬명**: `jikime-marketing-page-cro`

랜딩페이지 및 전환율 최적화 전문가로, 방문자를 고객으로 전환하는 전략을 제공합니다.

### 핵심 원칙

- **One Page, One Goal**: 페이지당 하나의 목표
- **Clarity Over Cleverness**: 명확함이 창의성보다 우선
- **Evidence Over Opinions**: 데이터 기반 의사결정

### CRO 매트릭스

```
┌─────────────────────────────────────────────────────────────────┐
│  CONVERSION OPTIMIZATION MATRIX                                  │
├─────────────────────────────────────────────────────────────────┤
│  RELEVANCE  ──────────────────────────────────────▶ FRICTION    │
│  (Match)                                           (Barriers)   │
│       │                                                 │       │
│       ▼                                                 ▼       │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐     │
│  │ Message │───▶│ Value   │───▶│ Trust   │───▶│ Action  │     │
│  │ Match   │    │ Prop    │    │ Signals │    │ Clarity │     │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

### 활용 예시

```bash
# 랜딩페이지 CRO 분석
/skill jikime-marketing-page-cro "분석해줘: https://example.com/landing"

# 전환율 개선 제안
/skill jikime-marketing-page-cro "SaaS 랜딩페이지 체크리스트"
```

---

## 2. Copywriting

**스킬명**: `jikime-marketing-copywriting`

전환 중심 카피라이팅 전문가로, 사용자 행동을 유도하는 카피를 작성합니다.

### 핵심 프레임워크

| 프레임워크 | 구조 | 용도 |
|------------|------|------|
| **PAS** | Problem → Agitate → Solution | 페인 포인트 강조 |
| **AIDA** | Attention → Interest → Desire → Action | 전통적 세일즈 |
| **BAB** | Before → After → Bridge | 변화 스토리 |
| **4Ps** | Promise → Picture → Proof → Push | 완전한 설득 |

### 헤드라인 공식

```
[숫자] + [형용사] + [키워드] + [약속] + [시간]
예: "7가지 검증된 전략으로 전환율을 2주 안에 2배로"
```

### 활용 예시

```bash
# 헤드라인 생성
/skill jikime-marketing-copywriting "SaaS 프로젝트 관리 도구 헤드라인 5개"

# CTA 최적화
/skill jikime-marketing-copywriting "무료 체험 CTA 버튼 카피"
```

---

## 3. Pricing

**스킬명**: `jikime-marketing-pricing`

가격 전략 및 리서치 전문가로, 최적의 가격 포인트와 구조를 결정합니다.

### Van Westendorp 분석

```
┌─────────────────────────────────────────────────────────────────┐
│  VAN WESTENDORP PRICE SENSITIVITY METER                         │
├─────────────────────────────────────────────────────────────────┤
│  100%│                                                          │
│      │    Too Cheap    ╲      ╱    Too Expensive               │
│      │                  ╲    ╱                                  │
│      │                   ╲  ╱                                   │
│   50%│                    ╲╱ ← Optimal Price Point             │
│      │                    ╱╲                                    │
│      │                   ╱  ╲                                   │
│      │    Bargain      ╱    ╲    Expensive                     │
│    0%└──────────────────────────────────────────────────────   │
│       $0                    Price                    $∞         │
└─────────────────────────────────────────────────────────────────┘
```

### 4가지 질문

1. 너무 비싸서 고려하지 않을 가격?
2. 비싸지만 고려할 수 있는 가격?
3. 좋은 거래라고 느끼는 가격?
4. 너무 싸서 품질이 의심되는 가격?

### 활용 예시

```bash
# 가격 전략 수립
/skill jikime-marketing-pricing "B2B SaaS 3-tier 가격 구조"

# 가격 리서치 설계
/skill jikime-marketing-pricing "Van Westendorp 설문 설계"
```

---

## 4. Psychology

**스킬명**: `jikime-marketing-psychology`

70개 이상의 마케팅 심리 모델을 제공하여 사용자 행동을 이해하고 영향을 줍니다.

### 6가지 카테고리

| 카테고리 | 모델 수 | 대표 모델 |
|----------|---------|-----------|
| **Persuasion** | 15+ | Social Proof, Reciprocity, Authority |
| **Cognitive Biases** | 20+ | Anchoring, Loss Aversion, Framing |
| **Decision Making** | 10+ | Paradox of Choice, Default Effect |
| **Motivation** | 10+ | Self-Determination, Goal Gradient |
| **Memory & Attention** | 10+ | Peak-End Rule, Serial Position |
| **Social Dynamics** | 10+ | Bandwagon, In-Group Bias |

### 핵심 모델 예시

```
┌─────────────────────────────────────────────────────────────────┐
│  ANCHORING EFFECT                                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  "Was $299"  ←── Anchor (높은 기준점 설정)                       │
│  "Now $99"   ←── Target (상대적으로 저렴해 보임)                 │
│                                                                  │
│  적용: 가격표, 할인, 비교 차트                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 활용 예시

```bash
# 심리 원칙 적용
/skill jikime-marketing-psychology "랜딩페이지에 Social Proof 적용"

# 설득 전략
/skill jikime-marketing-psychology "이탈 방지를 위한 Loss Aversion"
```

---

## 5. SEO

**스킬명**: `jikime-marketing-seo`

SEO 감사 및 최적화 전문가로, 검색 엔진 가시성을 향상시킵니다.

### SEO 우선순위 프레임워크

| 우선순위 | 영역 | 체크 항목 |
|----------|------|-----------|
| **P0 Critical** | 인덱싱 | robots.txt, sitemap, canonical |
| **P1 High** | 기술 SEO | 페이지 속도, 모바일, HTTPS |
| **P2 Medium** | 온페이지 | 타이틀, 메타, 헤딩, 콘텐츠 |
| **P3 Low** | 고급 | 스키마, 내부링크, UX |

### 기술 SEO 체크리스트

```
□ robots.txt 접근성
□ XML sitemap 존재 및 제출
□ canonical URL 설정
□ 페이지 속도 (LCP < 2.5s)
□ 모바일 친화성
□ HTTPS 적용
□ 구조화된 데이터 (Schema.org)
```

### 활용 예시

```bash
# SEO 감사
/skill jikime-marketing-seo "전체 SEO 감사 체크리스트"

# 기술 SEO 개선
/skill jikime-marketing-seo "Core Web Vitals 최적화 전략"
```

---

## 6. A/B Test

**스킬명**: `jikime-marketing-ab-test`

A/B 테스트 설계 및 분석 전문가로, 통계적으로 유의미한 실험을 설계합니다.

### 가설 프레임워크

```
┌─────────────────────────────────────────────────────────────────┐
│  HYPOTHESIS TEMPLATE                                             │
├─────────────────────────────────────────────────────────────────┤
│  Because [관찰/데이터],                                          │
│  we believe [변경사항]                                           │
│  will cause [예상 결과]                                          │
│  for [대상 사용자].                                              │
│  We'll know this is true when [측정 지표].                      │
└─────────────────────────────────────────────────────────────────┘
```

### 샘플 사이즈 가이드

| 기준 전환율 | 10% 개선 | 20% 개선 | 50% 개선 |
|-------------|----------|----------|----------|
| 1% | 150K/variant | 39K/variant | 6K/variant |
| 3% | 47K/variant | 12K/variant | 2K/variant |
| 5% | 27K/variant | 7K/variant | 1.2K/variant |
| 10% | 12K/variant | 3K/variant | 550/variant |

### 활용 예시

```bash
# 테스트 설계
/skill jikime-marketing-ab-test "CTA 버튼 색상 A/B 테스트 설계"

# 결과 분석
/skill jikime-marketing-ab-test "테스트 결과 해석 가이드"
```

---

## 7. Email

**스킬명**: `jikime-marketing-email`

이메일 마케팅 및 자동화 전문가로, 효과적인 이메일 시퀀스를 설계합니다.

### 시퀀스 타입

```
┌─────────────────────────────────────────────────────────────────┐
│                    EMAIL SEQUENCE TYPES                          │
├─────────────────────────────────────────────────────────────────┤
│  WELCOME         NURTURE          RE-ENGAGE        ONBOARDING   │
│  ───────         ───────          ─────────        ──────────   │
│  3-7 emails      5-10 emails      3-5 emails       5-10 emails  │
│  Post-signup     Pre-sale         30-60d inactive  Product user │
│  Build trust     Educate          Win back         Activate     │
└─────────────────────────────────────────────────────────────────┘
```

### Welcome 시퀀스 템플릿

| 이메일 | 타이밍 | 목적 |
|--------|--------|------|
| 1 | 즉시 | 약속 전달, 기대 설정 |
| 2 | Day 1-2 | 빠른 성공 경험 |
| 3 | Day 3-4 | 감정적 연결 (Why) |
| 4 | Day 5-6 | 소셜 프루프 |
| 5 | Day 7-8 | 반론 처리 |
| 6 | Day 9-11 | 기능 발견 |
| 7 | Day 12-14 | 전환 유도 |

### 활용 예시

```bash
# 웰컴 시퀀스
/skill jikime-marketing-email "SaaS 웰컴 이메일 7개 시퀀스"

# 재참여 캠페인
/skill jikime-marketing-email "휴면 고객 재활성화 이메일"
```

---

## 8. Launch

**스킬명**: `jikime-marketing-launch`

제품 런칭 및 Go-to-Market 전략 전문가로, 성공적인 런칭을 계획합니다.

### ORB 프레임워크

```
┌─────────────────────────────────────────────────────────────────┐
│                    CHANNEL STRATEGY (ORB)                        │
├─────────────────────────────────────────────────────────────────┤
│  OWNED              RENTED              BORROWED                │
│  ──────             ──────              ────────                │
│  You control        You rent visibility  You tap others'        │
│  the channel        from platforms       audiences              │
│                                                                  │
│  • Email list       • Social media       • Guest content        │
│  • Blog             • App stores         • Podcast interviews   │
│  • Podcast          • YouTube            • Collaborations       │
│  • Community        • Reddit             • Influencer partners  │
│  • Website          • Marketplaces       • Speaking events      │
│                                                                  │
│  [Most Valuable] ────────────────────── [Speed but Volatile]    │
└─────────────────────────────────────────────────────────────────┘
```

### 5단계 런칭

| 단계 | 목표 | 핵심 액션 |
|------|------|-----------|
| **1. Internal** | 핵심 기능 검증 | 5-10명 친한 테스터 |
| **2. Alpha** | 첫 외부 검증 | 랜딩페이지, 웨이트리스트 |
| **3. Beta** | 버즈 빌딩 | 티저 마케팅, 피드백 |
| **4. Early Access** | 스케일 검증 | 점진적 초대, PMF 설문 |
| **5. Full Launch** | 최대 가시성 | 오픈 사인업, 전체 발표 |

### 활용 예시

```bash
# 런칭 플랜
/skill jikime-marketing-launch "B2B SaaS 5단계 런칭 플랜"

# Product Hunt 전략
/skill jikime-marketing-launch "Product Hunt 런칭 체크리스트"
```

---

## 9. Analytics

**스킬명**: `jikime-marketing-analytics`

분석 추적 구현 전문가로, GA4, GTM, 이벤트 추적을 설정합니다.

### 이벤트 명명 규칙

```javascript
// Object-Action 형식 (권장)
signup_completed
button_clicked
form_submitted
article_read

// Category_Object_Action (복잡한 제품)
checkout_payment_completed
blog_article_viewed
onboarding_step_completed
```

### UTM 파라미터 전략

| 파라미터 | 목적 | 예시 |
|----------|------|------|
| `utm_source` | 트래픽 출처 | google, facebook, newsletter |
| `utm_medium` | 마케팅 매체 | cpc, email, social |
| `utm_campaign` | 캠페인명 | spring_sale, product_launch |
| `utm_content` | 버전 구분 | hero_cta, sidebar_link |
| `utm_term` | 유료 키워드 | running+shoes |

### GA4 구현 체크리스트

```
□ 플랫폼별 데이터 스트림 (web, iOS, Android)
□ Enhanced Measurement 활성화
□ 권장 이벤트 (Google 명명 규칙)
□ 커스텀 이벤트 (비즈니스 특화)
□ 전환 이벤트 마킹 (Admin > Events)
□ 커스텀 디멘션 설정
```

### 활용 예시

```bash
# 추적 계획
/skill jikime-marketing-analytics "SaaS 이벤트 추적 계획"

# GTM 설정
/skill jikime-marketing-analytics "GTM 컨테이너 구조 설계"
```

---

## 10. Onboarding

**스킬명**: `jikime-marketing-onboarding`

사용자 온보딩 및 활성화 전문가로, Time-to-Value를 최소화합니다.

### 핵심 원칙

- **Time-to-Value Is Everything**: 가입과 첫 가치 사이 모든 단계 제거
- **One Goal Per Session**: 한 번에 모든 것을 가르치지 않기
- **Do, Don't Show**: 인터랙티브 > 튜토리얼, 하기 > 배우기
- **Progress Creates Motivation**: 진행 상황 표시, 완료 축하

### 제품별 Aha Moment

| 제품 타입 | 전형적인 Aha Moment |
|-----------|---------------------|
| **프로젝트 관리** | 첫 프로젝트 생성 + 팀원 추가 |
| **분석 도구** | 추적 설치 + 첫 리포트 확인 |
| **디자인 도구** | 첫 디자인 생성 + 내보내기/공유 |
| **협업 도구** | 첫 팀원 초대 |
| **마켓플레이스** | 첫 거래 완료 |

### 온보딩 체크리스트 가이드라인

| 요소 | 가이드라인 |
|------|-----------|
| **항목 수** | 3-7개 (압도하지 않게) |
| **순서** | 가장 영향력 있는 것 먼저 |
| **시작** | 빠른 성공 경험으로 |
| **진행률** | 바 또는 완료 % |
| **완료** | 축하 순간 |
| **탈출** | 해제 옵션 (가두지 않기) |

### 활용 예시

```bash
# 온보딩 플로우
/skill jikime-marketing-onboarding "SaaS 온보딩 체크리스트 설계"

# 활성화 개선
/skill jikime-marketing-onboarding "이탈 사용자 재참여 전략"
```

---

## 스킬 연계 가이드

마케팅 스킬들은 서로 연계하여 사용할 때 더 강력합니다:

### 런칭 캠페인

```
launch → email → analytics → ab-test
  │        │         │          │
  │        │         │          └── 랜딩페이지 테스트
  │        │         └── 캠페인 추적
  │        └── 런칭 시퀀스
  └── 5단계 런칭 플랜
```

### 전환 최적화

```
page-cro → psychology → copywriting → ab-test
    │          │            │            │
    │          │            │            └── 변경 검증
    │          │            └── 카피 개선
    │          └── 심리 원칙 적용
    └── 페이지 분석
```

### 사용자 여정

```
launch → onboarding → email → analytics
   │          │          │        │
   │          │          │        └── 행동 추적
   │          │          └── 라이프사이클 이메일
   │          └── 활성화 최적화
   └── 첫 사용자 획득
```

---

## 트리거 키워드

각 스킬은 다음 키워드로 자동 활성화됩니다:

| 스킬 | 한국어 키워드 | 영문 키워드 |
|------|---------------|-------------|
| page-cro | 랜딩페이지, 전환율, CRO | landing page, conversion, CRO |
| copywriting | 카피, 헤드라인, CTA | copy, headline, CTA |
| pricing | 가격, 요금제, 프라이싱 | pricing, tier, plan |
| psychology | 심리, 설득, 행동 | psychology, persuasion, bias |
| seo | SEO, 검색최적화, 메타 | SEO, search optimization |
| ab-test | AB테스트, 실험, 가설 | A/B test, experiment |
| email | 이메일, 시퀀스, 드립 | email sequence, drip |
| launch | 런칭, 출시, 베타 | launch, Product Hunt |
| analytics | 분석, 추적, GA4 | analytics, tracking, GTM |
| onboarding | 온보딩, 활성화 | onboarding, activation |

---

## 참고 자료

- **원본 저장소**: [marketingskills](https://github.com/coreyhaines/marketingskills)
- **라이선스**: MIT License
- **원저자**: Corey Haines

---

*Version: 1.0.0 | Last Updated: 2026-01-25*
