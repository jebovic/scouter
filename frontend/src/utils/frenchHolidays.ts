export interface FrenchHoliday {
  date: string // YYYY-MM-DD format
  name: string // French name
  nameEn: string // English name
  type: 'public' | 'school_break'
  storesLikelyClosed: boolean
}

/**
 * Calculate Easter Sunday using the Anonymous Gregorian algorithm
 * Returns the date as a Date object
 */
function getEasterDate(year: number): Date {
  // Anonymous Gregorian algorithm
  const a = year % 19
  const b = Math.floor(year / 100)
  const c = year % 100
  const d = Math.floor(b / 4)
  const e = b % 4
  const f = Math.floor((b + 8) / 25)
  const g = Math.floor((b - f + 1) / 3)
  const h = (19 * a + b - d - g + 15) % 30
  const i = Math.floor(c / 4)
  const k = c % 4
  const l = (32 + 2 * e + 2 * i - h - k) % 7
  const m = Math.floor((a + 11 * h + 22 * l) / 451)
  const month = Math.floor((h + l - 7 * m + 114) / 31)
  const day = ((h + l - 7 * m + 114) % 31) + 1

  return new Date(year, month - 1, day)
}

/**
 * Format a date as YYYY-MM-DD string
 */
function formatDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Add days to a date and return new date
 */
function addDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}

/**
 * Get all French public holidays and school breaks for a given year
 */
export function getFrenchHolidays(year: number): FrenchHoliday[] {
  const holidays: FrenchHoliday[] = []

  // Fixed public holidays
  holidays.push({
    date: `${year}-01-01`,
    name: 'Jour de l\'An',
    nameEn: 'New Year\'s Day',
    type: 'public',
    storesLikelyClosed: false,
  })

  holidays.push({
    date: `${year}-05-01`,
    name: 'Fête du Travail',
    nameEn: 'Labour Day',
    type: 'public',
    storesLikelyClosed: true,
  })

  holidays.push({
    date: `${year}-05-08`,
    name: 'Victoire 1945',
    nameEn: 'Victory in Europe Day',
    type: 'public',
    storesLikelyClosed: false,
  })

  holidays.push({
    date: `${year}-07-14`,
    name: 'Fête Nationale',
    nameEn: 'Bastille Day',
    type: 'public',
    storesLikelyClosed: false,
  })

  holidays.push({
    date: `${year}-08-15`,
    name: 'Assomption',
    nameEn: 'Assumption of Mary',
    type: 'public',
    storesLikelyClosed: false,
  })

  holidays.push({
    date: `${year}-11-01`,
    name: 'Toussaint',
    nameEn: 'All Saints\' Day',
    type: 'public',
    storesLikelyClosed: true,
  })

  holidays.push({
    date: `${year}-11-11`,
    name: 'Armistice',
    nameEn: 'Armistice Day',
    type: 'public',
    storesLikelyClosed: false,
  })

  holidays.push({
    date: `${year}-12-25`,
    name: 'Noël',
    nameEn: 'Christmas',
    type: 'public',
    storesLikelyClosed: true,
  })

  // Easter-based moveable holidays
  const easterDate = getEasterDate(year)

  // Easter Monday = Easter + 1 day
  const easterMonday = addDays(easterDate, 1)
  holidays.push({
    date: formatDate(easterMonday),
    name: 'Lundi de Pâques',
    nameEn: 'Easter Monday',
    type: 'public',
    storesLikelyClosed: false,
  })

  // Ascension = Easter + 39 days
  const ascension = addDays(easterDate, 39)
  holidays.push({
    date: formatDate(ascension),
    name: 'Ascension',
    nameEn: 'Ascension',
    type: 'public',
    storesLikelyClosed: false,
  })

  // Pentecost Monday = Easter + 50 days
  const pentecostMonday = addDays(easterDate, 50)
  holidays.push({
    date: formatDate(pentecostMonday),
    name: 'Lundi de Pentecôte',
    nameEn: 'Whit Monday',
    type: 'public',
    storesLikelyClosed: false,
  })

  // Sort by date
  return holidays.sort((a, b) => a.date.localeCompare(b.date))
}

/**
 * Get upcoming holidays from today (at least 'count' holidays in the future)
 * Returns at most 'count' items
 */
export function getUpcomingHolidays(count = 5): FrenchHoliday[] {
  const today = new Date()
  const year = today.getFullYear()
  const todayStr = today.toISOString().slice(0, 10)

  // Get holidays from current year and next year
  const allHolidays = [...getFrenchHolidays(year), ...getFrenchHolidays(year + 1)]

  // Filter for future holidays (>= today) and limit to count
  return allHolidays.filter((h) => h.date >= todayStr).slice(0, count)
}

/**
 * Check if a date is a French holiday, return the holiday object or undefined
 * Works with local dates (accounts for timezone)
 */
export function isHoliday(date: Date): FrenchHoliday | undefined {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const dateStr = `${year}-${month}-${day}`
  return getFrenchHolidays(year).find((h) => h.date === dateStr)
}
