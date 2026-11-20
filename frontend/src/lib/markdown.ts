import type { PRDContent } from "@/lib/api";

export type PRDFieldKind = "text" | "list" | "requirements";

export interface PRDField {
  key: keyof PRDContent;
  label: string;
  kind: PRDFieldKind;
}

export interface RequirementItem {
  id: string;
  description: string;
  priority: string;
  acceptanceCriteria: string[];
}

export const PRD_FIELDS: PRDField[] = [
  { key: "executiveSummary", label: "Ringkasan eksekutif", kind: "text" },
  { key: "productOverview", label: "Gambaran produk", kind: "text" },
  { key: "problemStatement", label: "Pernyataan masalah", kind: "text" },
  { key: "visionAndGoals", label: "Visi dan tujuan", kind: "list" },
  { key: "targetAudience", label: "Target pengguna", kind: "list" },
  { key: "useCases", label: "Use cases", kind: "list" },
  { key: "userStories", label: "User stories", kind: "requirements" },
  { key: "functionalRequirements", label: "Kebutuhan fungsional", kind: "requirements" },
  { key: "nonFunctionalRequirements", label: "Kebutuhan non-fungsional", kind: "requirements" },
  { key: "featureSpecifications", label: "Spesifikasi fitur", kind: "list" },
  { key: "userJourney", label: "Perjalanan pengguna", kind: "list" },
  { key: "scope", label: "Cakupan", kind: "list" },
  { key: "outOfScope", label: "Di luar cakupan", kind: "list" },
  { key: "businessRules", label: "Aturan bisnis", kind: "list" },
  { key: "dataRequirements", label: "Kebutuhan data", kind: "list" },
  { key: "apiRequirements", label: "Kebutuhan API", kind: "list" },
  { key: "architecture", label: "Arsitektur tingkat tinggi", kind: "text" },
  { key: "securityAndPrivacy", label: "Keamanan dan privasi", kind: "list" },
  { key: "successMetrics", label: "Metrik keberhasilan", kind: "list" },
  { key: "dependenciesAndConstraints", label: "Dependensi dan batasan", kind: "list" },
  { key: "risks", label: "Risiko dan mitigasi", kind: "list" },
  { key: "milestones", label: "Milestone dan roadmap", kind: "list" },
  { key: "openQuestions", label: "Pertanyaan terbuka", kind: "list" },
  { key: "assumptions", label: "Asumsi", kind: "list" },
];

export function valueToText(value: unknown, kind: PRDFieldKind): string {
  if (Array.isArray(value)) {
    return value
      .map((item) =>
        kind === "requirements" && item && typeof item === "object"
          ? (item as { description?: string }).description || ""
          : String(item)
      )
      .filter(Boolean)
      .join("\n");
  }
  return typeof value === "string" ? value : "";
}

export function prdToMarkdown(title: string, content?: PRDContent | null): string {
  if (!content) return `# ${title}`;

  const core = PRD_FIELDS.map((field) => {
    const value = content[field.key];
    if (!value || (Array.isArray(value) && value.length === 0)) return "";
    const text = Array.isArray(value)
      ? value
          .map((item) => {
            if (field.kind === "requirements" && item && typeof item === "object") {
              const req = item as RequirementItem;
              const criteria = req.acceptanceCriteria?.length
                ? `\n${req.acceptanceCriteria.map((c) => `  - [ ] ${c}`).join("\n")}`
                : "";
              return `- **${req.id || ""}** [${req.priority || "P1"}]: ${req.description || ""}${criteria}`;
            }
            return `- ${String(item)}`;
          })
          .join("\n")
      : String(value);
    return `## ${field.label}\n\n${text}`;
  })
    .filter(Boolean)
    .join("\n\n");

  const extra = Array.isArray(content.customSections)
    ? (content.customSections as Array<{ title: string; body: string }>)
        .map((section) => `## ${section.title}\n\n${section.body || ""}`)
        .join("\n\n")
    : "";

  return `# ${title}\n\n${[core, extra].filter(Boolean).join("\n\n")}`;
}
