# GoTest Agent — UI/UX Audit & Redesign Plan

**Date:** 2026-08-01  
**Status:** Critical Review Required  
**Target:** Production-quality SaaS UX (Linear/Vercel/Stripe level)

---

## 🔍 EXECUTIVE SUMMARY

The current UI/UX exhibits clear signs of AI-generated design rather than professional product craftsmanship. While functionally complete, the interface requires a comprehensive redesign to achieve modern SaaS quality standards.

### Major Issues Identified

1. **Visual Clutter** — Excessive borders, shadows, and card patterns throughout
2. **Weak Typography Hierarchy** — Inconsistent sizing (9px to 26px), poor rhythm
3. **Information Overload** — Dense layouts without proper breathing room
4. **Dated Components** — Tab patterns, badge styles, and generic elements
5. **Inconsistent Design System** — Mixed spacing tokens, unclear rules
6. **Lack of Focal Points** — No clear primary actions or visual anchors

---

## 📋 DETAILED AUDIT BY PAGE

### Global/Layout Issues

#### ❌ Current Problems

| Issue | Severity | Evidence | Impact |
|-------|----------|----------|--------|
| **Border Fatigue** | High | Every component has borders (`border-[var(--border)]`) | Visual noise, aged appearance |
| **Shadow Abuse** | High | Multiple shadow levels (`--shadow-xs`, `--shadow-sm`, `--shadow-md`) used indiscriminately | Muddy appearance, performance cost |
| **Rounded Corner Overuse** | Medium | Every button/card uses `rounded-[var(--radius)]` | Generic "Bootstrap-like" feel |
| **Sidebar Width** | Low | Fixed `w-[220px]` feels cramped on wide screens | Poor utilization of desktop space |
| **Navigation Density** | Medium | 13+ nav items crammed into sidebar | Cognitive overload |

#### ✅ Recommended Fixes

```css
/* Modern SaaS: Remove most borders */
.card { border: none; background: transparent; }
.section { border-bottom: 1px solid var(--border); } /* Only horizontal dividers */

/* Minimal shadows for depth only */
.shadow-subtle { box-shadow: 0 1px 2px rgba(0,0,0,0.02); }
.shadow-float { box-shadow: 0 4px 12px rgba(0,0,0,0.05); }

/* Consistent radius */
.radius-sm { border-radius: 6px; }
.radius-md { border-radius: 12px; }
.radius-lg { border-radius: 16px; }
```

---

### Page-by-Page Audit

#### 1. Login Page (`/login`)

**Current State:**
```tsx
// ❌ Problematic patterns
<div className="bg-[var(--bg-card)] rounded-xl border border-[var(--border)] p-6">
  <label className="uppercase tracking-wider">{/* Too aggressive */}</label>
  <input className="rounded-lg border..." /> {/* Border-heavy */}
  <button className="hover:brightness-110">...</button> {/* Forced hover state */}
</div>
```

**Issues:**
- ❌ Card with border creates artificial separation
- ❌ "API Key" label too aggressive with uppercase/tracking
- ❌ Hover brightness manipulation feels forced
- ❌ Form padding (p-6) inconsistent with platform standard

**Recommended Design:**
```tsx
// ✅ Modern approach
<form className="space-y-4 max-w-[340px] mx-auto">
  <h1 className="text-2xl font-semibold tracking-tight">GoTest Agent</h1>
  <p className="text-[13px] text-muted-foreground">Sign in to your workspace</p>
  
  <div className="space-y-2">
    <label htmlFor="api-key" className="text-sm font-medium block">Access Token</label>
    <input 
      id="api-key"
      className="w-full px-3 py-2 text-sm bg-background border border-input rounded-md focus:ring-2 focus:ring-ring"
      placeholder="Enter your access token"
    />
  </div>
  
  <button className="w-full py-2 text-sm font-medium bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
    Continue
  </button>
</form>
```

**Key Improvements:**
- ✅ No border on container
- ✅ Smaller form width for focus (max-w-340)
- ✅ Cleaner typography (no uppercase labels)
- ✅ Proper focus ring system
- ✅ Button color using semantic variables

---

#### 2. Dashboard/Overview (`/`)

**Current State:**
```tsx
// ❌ Problematic patterns
<div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
  <StatCard label="Total" value={total} icon={<Activity />} />
  <StatCard label="Pass Rate" value={`${passRate}%`} icon={<CheckCircle2 />} />
</div>

<Section title="🔴 Running Now">
  {activeRuns.map((r) => (
    <Link className="block p-3 rounded-[var(--radius-sm)] border border-[var(--warning)]/20 bg-[var(--warning-bg)]">
      {/* Content */}
    </Link>
  ))}
</Section>
```

**Issues:**
- ❌ **Emoji in headings** (`🔴 Running Now`) — unprofessional
- ❌ **Every stat card bordered** — unnecessary visual noise
- ❌ **Icon on every stat** — wastes space, adds clutter
- ❌ **Warning backgrounds** on active runs — creates alarmist tone
- ❌ **Section titles** too long, compete with content
- ❌ **Multiple CTA buttons** — Create Test appears 3x per page

**Recommended Design:**
```tsx
// ✅ Modern approach
<div className="space-y-8">
  {/* Hero Section */}
  <div className="flex items-center justify-between">
    <div>
      <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
      <p className="text-sm text-muted-foreground mt-1">
        {activeCount > 0 
          ? `${activeCount} test${activeCount === 1 ? ' run' : 's'} currently executing`
          : 'No active tests'
        }
      </p>
    </div>
    <Button variant="primary">New Run</Button>
  </div>
  
  {/* Stats Row - Clean & Simple */}
  <div className="grid grid-cols-4 gap-4">
    <Metric value={totalCount} label="Total Runs" trend="+12%" />
    <Metric value={passRate + '%'} label="Pass Rate" trend="+3% this week" />
    <Metric value={failedCount} label="Failures" trend="-2% vs last week" />
    <Metric value={avgDuration} label="Avg Duration" trend="15s faster" />
  </div>
  
  {/* Active Runs Table */}
  {activeCount > 0 && (
    <div className="border-b pb-4">
      <h2 className="text-sm font-medium mb-3">Currently Executing</h2>
      <Table data={activeRuns}>
        <Column accessor="name" header="Run Name" />
        <Column accessor="status" header="Status" render={statusBadge} />
        <Column accessor="progress" header="Progress" render={progressBar} />
        <Column accessor="age" header="Age" render={timeAgo} />
      </Table>
    </div>
  )}
  
  {/* Recommendations */}
  <div className="grid grid-cols-2 gap-4">
    <RecommendationCard item={topRec} />
    <RiskCard risk={highestRisk} />
  </div>
</div>
```

**Key Principles:**
1. **Remove emoji** from all headings
2. **Tables over cards** for lists (better scanning)
3. **Trend indicators** on metrics instead of icons
4. **Single primary CTA** positioned consistently
5. **Horizontal dividers** instead of bordered sections

---

#### 3. Runs List (`/runs`)

**Current State:**
```tsx
// ❌ Problematic patterns
<button className="w-full flex items-center gap-4 px-4 py-3 hover:bg-[var(--bg-hover)]">
  <StatusBadge state={r.state} />
  <span className="font-mono text-xs">{r.id.slice(0, 8)}</span>
  <span className="truncate">{r.requirements}</span>
  {run_result && <span>{passed}✓ {failed}✗</span>}
  <span>{timeAgo(r.created_at)}</span>
</button>
```

**Issues:**
- ❌ **Button as row** — poor semantics, confusing interaction model
- ❌ **Chevrons/arrows** indicating expandable but no visual feedback
- ❌ **Mixed alignment** — some right-aligned, some left
- ❌ **Status badges** everywhere create color pollution
- ❌ **Monospace ID** takes equal weight as human-readable name
- ❌ **Tab filtering** uses old-school underline pattern

**Recommended Design:**
```tsx
// ✅ Modern approach
<FilterGroup defaultValue="all">
  <FilterItem value="all">All Runs ({count.all})</FilterItem>
  <FilterItem value="active">Active ({count.active})</FilterItem>
  <FilterItem value="passed">Passed ({count.passed})</FilterItem>
  <FilterItem value="failed">Failed ({count.failed})</FilterItem>
</FilterGroup>

<List>
  {runs.map(run => (
    <ListItem href={`/runs/${run.id}`}>
      <ListItemText>
        <Title>{run.requirements || 'Untitled'}</Title>
        <Subtitle>ID: {run.id.slice(0,8)} • {timeAgo(run.createdAt)}</Subtitle>
      </ListItemText>
      <ListMeta>
        <StatusBadge status={run.status} size="sm" />
        <Text variant="muted">{run.result?.summary}</Text>
      </ListMeta>
    </ListItem>
  ))}
</List>
```

**Key Improvements:**
- ✅ **Semantic `<a>` tags** for list items
- ✅ **Hierarchical text** — Title > Subtitle > Meta
- ✅ **Smaller status badges** when in list context
- ✅ **Consistent right-side metadata** alignment
- ✅ **Cleaner filter pills** (modern segment control)

---

#### 4. Test Library (`/tests`)

**Current State:**
```tsx
// ❌ Problematic patterns
<Tab value={activeTab} onValueChange={setActiveTab}>
  <TabsList>
    <TabsTrigger value="ui">UI</TabsTrigger>
    <TabsTrigger value="api">API</TabsTrigger>
  </TabsList>
</Tab>

<MaintenanceAdvisor maintenance={maintenance} /> {/* Complex card */}
```

**Issues:**
- ❌ **Dual-tab interface** unnecessary complexity
- ❌ **Maintenance Advisor** card too dense, overwhelming
- ❌ **Severity badges** in advisor (high/medium/low) create panic tone
- ❌ **Proposal workflow** buried in secondary interactions
- ❌ **Run button** inline with each row causes accidental clicks

**Recommended Design:**
```tsx
<div className="space-y-4">
  {/* Primary Actions */}
  <div className="flex items-center justify-between">
    <SearchInput placeholder="Find tests..." />
    <Button>Generate Tests</Button>
  </div>
  
  {/* Simple Tab Switch */}
  <SegmentedControl
    options={[{ label: 'UI Tests', value: 'ui' }, { label: 'API Tests', value: 'api' }]}
    value={activeTab}
  />
  
  {/* Test List */}
  <TestList tests={filteredTests} />
  
  {/* Maintenance Panel - Collapsible */}
  {maintenance.length > 0 && (
    <CollapsiblePanel defaultOpen={false} summary={`${maintenance.length} maintenance items`}>
      <MaintenanceItems items={maintenance} />
    </CollapsiblePanel>
  )}
</div>
```

**Key Changes:**
- ✅ **Simpler tab control** (segmented button style)
- ✅ **Collapse maintenance panel** by default
- ✅ **Clearer CTA placement** (not inline with rows)
- ✅ **Better visual hierarchy** between primary and secondary actions

---

## 🎨 DESIGN SYSTEM RECOMMENDATIONS

### 1. Color Palette (Minimal)

```css
/* Replace 8-color chaos with semantic palette */
:root {
  /* Neutral scale - grayscale only */
  --bg-page: #fafafa;
  --bg-surface: #ffffff;
  --bg-elevated: #f8f9fa;
  --bg-hover: #f5f6f7;
  
  --text-primary: #111827;
  --text-secondary: #6b7280;
  --text-muted: #9ca3af;
  
  --border-default: #e5e7eb;
  --border-strong: #d1d5db;
  
  /* Single accent color */
  --accent: #2563eb;
  --accent-hover: #1d4ed8;
  --accent-light: #eff6ff;
  
  /* Status - minimal usage */
  --success: #059669;
  --warning: #d97706;
  --danger: #dc2626;
}
```

**Rationale:**
- Remove color variety → cleaner, more focused
- Grayscale backbone → better readability
- Accent only for CTAs and interactions

---

### 2. Spacing System (Consistent)

```css
:root {
  --space-1: 4px;   /* Tightest */
  --space-2: 8px;   /* Component internals */
  --space-3: 12px;  /* Small gaps */
  --space-4: 16px;  /* Standard */
  --space-5: 20px;  /* Large gaps */
  --space-6: 24px;  /* Section padding */
  --space-8: 32px;  /* Component padding */
  --space-12: 48px; /* Page padding */
}
```

**Usage Rules:**
- Everything multiples of 4px
- Section padding = 24px minimum
- Card/internal padding = 16px
- Text gaps = 8-12px

---

### 3. Typography Scale (Fewer Sizes)

```css
:root {
  --font-display: 28px / 36px;      /* Page headers only */
  --font-heading: 20px / 28px;      /* Section titles */
  --font-title: 16px / 24px;        /* Card titles */
  --font-body: 14px / 22px;         /* Default body */
  --font-caption: 12px / 18px;      /* Captions, timestamps */
  --font-mono: 13px / 20px;         /* Code, IDs */
}
```

**Size Limit:** Use maximum 6 type sizes throughout app

---

### 4. Shadow System (Minimal)

```css
:root {
  --shadow-none: none;
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.06);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.08);
}
```

**Usage Guidelines:**
- Only use shadows on floating elements (dialogs, dropdowns)
- Never use on bordered containers
- Prefer elevation via spacing instead of shadow

---

### 5. Radius Values (2 Options)

```css
:root {
  --radius-sm: 4px;   /* Buttons, inputs */
  --radius-md: 8px;   /* Cards, panels */
  /* Remove all other radius values */
}
```

---

## 🚀 IMPLEMENTATION PLAN

### Phase 1: Foundation (Day 1-2)

1. ✅ Update `globals.css` with new design tokens
2. ✅ Create new primitive components:
   - `Button` (variant-based: primary, secondary, ghost)
   - `Input` (clean, minimal styling)
   - `Card` (borderless by default)
   - `Table` (modern striped design)
   - `Badge` (minimal, contextual)

### Phase 2: Navigation (Day 2-3)

1. ✅ Redesigned Sidebar (wider, cleaner, grouped navigation)
2. ✅ Updated Header (simplified, consistent search)
3. ✅ Mobile navigation drawer

### Phase 3: Core Pages (Day 3-6)

Redesign in this order:
1. ✅ Login page (simplest entry point)
2. ✅ Dashboard (critical path)
3. ✅ Runs list (high-frequency use)
4. ✅ Test Library (complex workflow)

### Phase 4: Detail Pages (Day 6-8)

1. ✅ Run detail view
2. ✅ Project management pages
3. ✅ Settings & configuration
4. ✅ Alerts & monitoring

### Phase 5: Polish (Day 8-10)

1. ✅ Empty states (inspiring, not empty)
2. ✅ Loading states (engaging skeletons)
3. ✅ Error states (helpful, actionable)
4. ✅ Micro-interactions (subtle, purposeful)

---

## ✅ QUALITY CHECKLIST

Before considering the redesign complete, verify:

- [ ] Zero emoji in production UI
- [ ] Maximum 2 border radii used consistently
- [ ] All shadows under 0.1 opacity
- [ ] Type scale max 6 sizes
- [ ] Consistent 24px section padding
- [ ] Single primary CTA per viewport
- [ ] Tables instead of cards for lists
- [ ] Grayscale base with accent highlights only
- [ ] 0.8s transition on interactions max
- [ ] Focus rings visible on all interactive elements
- [ ] Screen reader friendly markup
- [ ] Keyboard navigable throughout
- [ ] Mobile responsive at all breakpoints

---

## 📊 BEFORE/AFTER METRICS

| Metric | Before | Target | Improvement |
|--------|--------|--------|-------------|
| **Bordered Elements** | ~80% | ~20% | 75% reduction |
| **Color Varieties** | 8+ | 4 | 50% reduction |
| **Type Sizes Used** | 12+ | 6 | 50% reduction |
| **Average Padding** | 12-20px | 24px | Better whitespace |
| **Visual Clarity** | ⭐⭐ | ⭐⭐⭐⭐⭐ | +150% |
| **Professional Feel** | ⭐⭐ | ⭐⭐⭐⭐⭐ | +150% |

---

## 🎯 SUCCESS CRITERIA

The redesign is successful when:

1. **Visually Identifiable** as modern SaaS product (Linear/Vercel-level)
2. **Zero "AI-generated" artifacts** remain
3. **Users report** feeling "this is professional software"
4. **Conversion rates improve** (if applicable)
5. **Support tickets decrease** (better UX = fewer confusions)

---

**Next Step:** Begin implementation starting with `globals.css` updates and primitive component creation.
