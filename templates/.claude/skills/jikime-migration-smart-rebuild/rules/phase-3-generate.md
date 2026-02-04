# Phase 3: Generate (코드 생성)

## 목표

mapping.json을 기반으로 Next.js 프론트엔드와 Java Spring Boot 백엔드 코드를 생성합니다.

## 실행

```bash
/jikime:smart-rebuild generate --mapping=./mapping.json --backend=java --frontend=nextjs
```

## 생성 전략

| 페이지 타입 | Frontend | Backend |
|------------|----------|---------|
| 정적 | Next.js 정적 페이지 | - |
| 동적 | Next.js + API 호출 | Java Entity/Repository/Controller |

## 정적 페이지 생성

### 입력

- 스크린샷 (UI 디자인 참고)
- HTML (텍스트, 이미지 추출)

### 출력

```tsx
// app/about/page.tsx
export default function AboutPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">회사 소개</h1>
      <div className="prose max-w-none">
        {/* HTML에서 추출한 콘텐츠 */}
        <p>
          저희 회사는 2010년에 설립되어...
        </p>
        <img
          src="/images/company.jpg"
          alt="회사 전경"
          className="w-full rounded-lg my-6"
        />
      </div>
    </div>
  );
}
```

### 생성 프롬프트

```
다음 스크린샷과 HTML을 참고하여 Next.js 페이지를 생성하세요.

입력:
- 스크린샷: captures/about.png (UI 레이아웃 참고)
- HTML: captures/about.html (텍스트, 이미지 URL 추출)

요구사항:
- Tailwind CSS 사용
- 반응형 디자인
- 접근성 고려 (alt, aria 속성)
- 이미지는 public/images/에 복사 후 참조
```

## 동적 페이지 생성

### Backend (Java Spring Boot)

#### Entity

```java
// src/main/java/com/example/entity/Member.java
@Entity
@Table(name = "members")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class Member {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String email;

    @Column(nullable = false, length = 100)
    private String name;

    @Enumerated(EnumType.STRING)
    private MemberStatus status;

    @CreatedDate
    private LocalDateTime createdAt;
}
```

#### Repository

```java
// src/main/java/com/example/repository/MemberRepository.java
public interface MemberRepository extends JpaRepository<Member, Long> {

    List<Member> findByStatus(MemberStatus status);

    Optional<Member> findByEmail(String email);

    @Query("SELECT m FROM Member m WHERE m.status = :status ORDER BY m.createdAt DESC")
    Page<Member> findActiveMembers(@Param("status") MemberStatus status, Pageable pageable);
}
```

#### Controller

```java
// src/main/java/com/example/controller/MemberController.java
@RestController
@RequestMapping("/api/members")
@RequiredArgsConstructor
public class MemberController {

    private final MemberRepository memberRepository;

    @GetMapping
    public ResponseEntity<List<Member>> getActiveMembers() {
        List<Member> members = memberRepository.findByStatus(MemberStatus.ACTIVE);
        return ResponseEntity.ok(members);
    }

    @GetMapping("/{id}")
    public ResponseEntity<Member> getMember(@PathVariable Long id) {
        return memberRepository.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<Member> createMember(@Valid @RequestBody MemberRequest request) {
        Member member = new Member();
        member.setEmail(request.getEmail());
        member.setName(request.getName());
        member.setStatus(MemberStatus.ACTIVE);

        Member saved = memberRepository.save(member);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }
}
```

### Frontend (Next.js)

#### Server Component (데이터 페칭)

```tsx
// app/members/page.tsx
import { MemberCard } from '@/components/member-card';

interface Member {
  id: number;
  email: string;
  name: string;
  status: string;
}

async function getMembers(): Promise<Member[]> {
  const res = await fetch(`${process.env.API_URL}/api/members`, {
    cache: 'no-store'
  });

  if (!res.ok) {
    throw new Error('Failed to fetch members');
  }

  return res.json();
}

export default async function MembersPage() {
  const members = await getMembers();

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">회원 목록</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {members.map((member) => (
          <MemberCard key={member.id} member={member} />
        ))}
      </div>

      {members.length === 0 && (
        <p className="text-gray-500 text-center py-8">
          등록된 회원이 없습니다.
        </p>
      )}
    </div>
  );
}
```

#### Client Component (인터랙션)

```tsx
// components/member-card.tsx
'use client';

import { useState } from 'react';

interface MemberCardProps {
  member: {
    id: number;
    email: string;
    name: string;
    status: string;
  };
}

export function MemberCard({ member }: MemberCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div
      className="bg-white rounded-lg shadow-md p-4 hover:shadow-lg transition-shadow cursor-pointer"
      onClick={() => setIsExpanded(!isExpanded)}
    >
      <h3 className="font-semibold text-lg">{member.name}</h3>
      <p className="text-gray-600 text-sm">{member.email}</p>

      {isExpanded && (
        <div className="mt-4 pt-4 border-t">
          <span className={`
            px-2 py-1 rounded text-xs
            ${member.status === 'ACTIVE'
              ? 'bg-green-100 text-green-800'
              : 'bg-gray-100 text-gray-800'}
          `}>
            {member.status}
          </span>
        </div>
      )}
    </div>
  );
}
```

## SQL → Entity 매핑

| SQL Type | Java Type | Notes |
|----------|-----------|-------|
| INT | Long | Auto-increment → @GeneratedValue |
| VARCHAR(n) | String | @Column(length = n) |
| TEXT | String | @Lob |
| DATETIME | LocalDateTime | @CreatedDate |
| ENUM | Enum | @Enumerated(EnumType.STRING) |
| BOOLEAN | Boolean | - |
| DECIMAL | BigDecimal | 금액 처리 |

## CLI 옵션

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--mapping` | mapping.json 경로 | (필수) |
| `--backend` | 백엔드 프레임워크 | `java` |
| `--frontend` | 프론트엔드 프레임워크 | `nextjs` |
| `--output-backend` | 백엔드 출력 디렉토리 | `./backend` |
| `--output-frontend` | 프론트엔드 출력 디렉토리 | `./frontend` |
| `--style` | CSS 프레임워크 | `tailwind` |

## 생성 파일 구조

```
output/
├── backend/
│   └── src/main/java/com/example/
│       ├── entity/
│       │   └── Member.java
│       ├── repository/
│       │   └── MemberRepository.java
│       └── controller/
│           └── MemberController.java
│
└── frontend/
    ├── app/
    │   ├── about/
    │   │   └── page.tsx (정적)
    │   └── members/
    │       └── page.tsx (동적)
    ├── components/
    │   └── member-card.tsx
    └── public/
        └── images/
            └── (추출된 이미지들)
```

## 다음 단계

→ [Troubleshooting](./troubleshooting.md)
