import * as fs from 'fs';
import * as path from 'path';

interface GenerateOptions {
  mappingFile: string;
  backend: string;
  frontend: string;
  outputBackend: string;
  outputFrontend: string;
  style?: string;
}

interface PageMapping {
  id: string;
  capture: {
    url: string;
    screenshot: string;
    html: string;
  };
  source: {
    file: string | null;
    type: 'static' | 'dynamic' | 'unknown';
    reason: string[];
  };
  database?: {
    queries: Array<{
      raw: string;
      table: string;
      type: string;
      columns?: string[];
    }>;
  };
  output: {
    backend?: {
      entity?: string;
      repository?: string;
      controller?: string;
      endpoint?: string;
    };
    frontend: {
      path: string;
      type: 'static-page' | 'dynamic-page';
      apiCalls?: string[];
    };
  };
}

interface Mapping {
  project: {
    name: string;
    sourceUrl: string;
    sourcePath: string;
  };
  pages: PageMapping[];
}

/**
 * 테이블 이름을 Entity 이름으로 변환
 */
function toEntityName(tableName: string): string {
  return tableName
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join('');
}

/**
 * 테이블 이름을 camelCase로 변환
 */
function toCamelCase(tableName: string): string {
  const entity = toEntityName(tableName);
  return entity.charAt(0).toLowerCase() + entity.slice(1);
}

/**
 * Java Entity 생성
 */
function generateJavaEntity(tableName: string, columns?: string[]): string {
  const entityName = toEntityName(tableName);

  return `package com.example.entity;

import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import java.time.LocalDateTime;

@Entity
@Table(name = "${tableName}")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ${entityName} {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    // TODO: Add columns based on database schema
    // Columns detected: ${columns?.join(', ') || 'unknown'}

    @Column(name = "created_at")
    private LocalDateTime createdAt;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;
}
`;
}

/**
 * Java Repository 생성
 */
function generateJavaRepository(tableName: string): string {
  const entityName = toEntityName(tableName);

  return `package com.example.repository;

import com.example.entity.${entityName};
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ${entityName}Repository extends JpaRepository<${entityName}, Long> {

    // TODO: Add custom query methods based on SQL analysis

}
`;
}

/**
 * Java Controller 생성
 */
function generateJavaController(tableName: string): string {
  const entityName = toEntityName(tableName);
  const varName = toCamelCase(tableName);

  return `package com.example.controller;

import com.example.entity.${entityName};
import com.example.repository.${entityName}Repository;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/${tableName}")
@RequiredArgsConstructor
public class ${entityName}Controller {

    private final ${entityName}Repository ${varName}Repository;

    @GetMapping
    public ResponseEntity<List<${entityName}>> getAll() {
        List<${entityName}> ${varName}s = ${varName}Repository.findAll();
        return ResponseEntity.ok(${varName}s);
    }

    @GetMapping("/{id}")
    public ResponseEntity<${entityName}> getById(@PathVariable Long id) {
        return ${varName}Repository.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<${entityName}> create(@RequestBody ${entityName} ${varName}) {
        ${entityName} saved = ${varName}Repository.save(${varName});
        return ResponseEntity.ok(saved);
    }

    @PutMapping("/{id}")
    public ResponseEntity<${entityName}> update(@PathVariable Long id, @RequestBody ${entityName} ${varName}) {
        if (!${varName}Repository.existsById(id)) {
            return ResponseEntity.notFound().build();
        }
        ${varName}.setId(id);
        ${entityName} updated = ${varName}Repository.save(${varName});
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id) {
        if (!${varName}Repository.existsById(id)) {
            return ResponseEntity.notFound().build();
        }
        ${varName}Repository.deleteById(id);
        return ResponseEntity.noContent().build();
    }
}
`;
}

/**
 * Next.js 정적 페이지 생성
 */
function generateStaticPage(pagePath: string, url: string): string {
  const pageName = path.basename(path.dirname(pagePath)) || 'Home';
  const titleCase = pageName.charAt(0).toUpperCase() + pageName.slice(1);

  return `// ${pagePath}
// Generated from: ${url}

export default function ${titleCase}Page() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">${titleCase}</h1>

      <div className="prose max-w-none">
        {/* TODO: Extract content from captured HTML */}
        <p>
          Content goes here...
        </p>
      </div>
    </div>
  );
}
`;
}

/**
 * Next.js 동적 페이지 생성
 */
function generateDynamicPage(
  pagePath: string,
  url: string,
  apiEndpoint: string,
  entityName: string
): string {
  const pageName = path.basename(path.dirname(pagePath)) || 'Items';
  const titleCase = pageName.charAt(0).toUpperCase() + pageName.slice(1);
  const varName = entityName.charAt(0).toLowerCase() + entityName.slice(1);

  return `// ${pagePath}
// Generated from: ${url}
// API: ${apiEndpoint}

interface ${entityName} {
  id: number;
  // TODO: Add fields based on entity
  createdAt: string;
  updatedAt: string;
}

async function get${entityName}s(): Promise<${entityName}[]> {
  const res = await fetch(\`\${process.env.API_URL}${apiEndpoint}\`, {
    cache: 'no-store',
  });

  if (!res.ok) {
    throw new Error('Failed to fetch ${varName}s');
  }

  return res.json();
}

export default async function ${titleCase}Page() {
  const ${varName}s = await get${entityName}s();

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">${titleCase}</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {${varName}s.map((${varName}) => (
          <div
            key={${varName}.id}
            className="bg-white rounded-lg shadow-md p-4 hover:shadow-lg transition-shadow"
          >
            <h3 className="font-semibold text-lg">ID: {${varName}.id}</h3>
            {/* TODO: Add more fields */}
          </div>
        ))}
      </div>

      {${varName}s.length === 0 && (
        <p className="text-gray-500 text-center py-8">
          No ${varName}s found.
        </p>
      )}
    </div>
  );
}
`;
}

/**
 * 코드 생성 메인 함수
 */
export async function generateCode(options: GenerateOptions): Promise<void> {
  const { mappingFile, backend, frontend, outputBackend, outputFrontend } = options;

  console.log('⚡ 코드 생성 시작');

  // 매핑 파일 로드
  if (!fs.existsSync(mappingFile)) {
    throw new Error(`Mapping file not found: ${mappingFile}`);
  }

  const mapping: Mapping = JSON.parse(fs.readFileSync(mappingFile, 'utf-8'));
  console.log(`📋 매핑 로드: ${mapping.pages.length}개 페이지`);

  // 출력 디렉토리 생성
  fs.mkdirSync(outputBackend, { recursive: true });
  fs.mkdirSync(outputFrontend, { recursive: true });

  let backendCount = 0;
  let frontendCount = 0;

  // 생성된 테이블 추적 (중복 방지)
  const generatedTables = new Set<string>();

  for (const page of mapping.pages) {
    // Backend 생성 (Java)
    if (backend === 'java' && page.output.backend && page.database?.queries) {
      for (const query of page.database.queries) {
        if (generatedTables.has(query.table)) continue;
        generatedTables.add(query.table);

        const entityName = toEntityName(query.table);

        // Entity
        const entityDir = path.join(outputBackend, 'src/main/java/com/example/entity');
        fs.mkdirSync(entityDir, { recursive: true });
        fs.writeFileSync(
          path.join(entityDir, `${entityName}.java`),
          generateJavaEntity(query.table, query.columns)
        );

        // Repository
        const repoDir = path.join(outputBackend, 'src/main/java/com/example/repository');
        fs.mkdirSync(repoDir, { recursive: true });
        fs.writeFileSync(
          path.join(repoDir, `${entityName}Repository.java`),
          generateJavaRepository(query.table)
        );

        // Controller
        const ctrlDir = path.join(outputBackend, 'src/main/java/com/example/controller');
        fs.mkdirSync(ctrlDir, { recursive: true });
        fs.writeFileSync(
          path.join(ctrlDir, `${entityName}Controller.java`),
          generateJavaController(query.table)
        );

        backendCount++;
        console.log(`   ✓ Backend: ${entityName} (Entity, Repository, Controller)`);
      }
    }

    // Frontend 생성 (Next.js)
    if (frontend === 'nextjs') {
      const frontendPath = page.output.frontend.path;
      const fullPath = path.join(outputFrontend, frontendPath);
      fs.mkdirSync(path.dirname(fullPath), { recursive: true });

      if (page.output.frontend.type === 'static-page') {
        fs.writeFileSync(fullPath, generateStaticPage(frontendPath, page.capture.url));
      } else {
        const apiEndpoint = page.output.frontend.apiCalls?.[0] || '/api/items';
        const table = page.database?.queries?.[0]?.table || 'Item';
        const entityName = toEntityName(table);
        fs.writeFileSync(
          fullPath,
          generateDynamicPage(frontendPath, page.capture.url, apiEndpoint, entityName)
        );
      }

      frontendCount++;
      console.log(`   ✓ Frontend: ${frontendPath}`);
    }
  }

  console.log(`\n✅ 코드 생성 완료!`);
  console.log(`📦 Backend: ${backendCount}개 엔티티`);
  console.log(`🎨 Frontend: ${frontendCount}개 페이지`);
  console.log(`📁 Backend: ${outputBackend}`);
  console.log(`📁 Frontend: ${outputFrontend}`);
}
