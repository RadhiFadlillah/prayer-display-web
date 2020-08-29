export function isValidDate(d) {
	return d instanceof Date && !isNaN(d);
}

export function isoTimeString(d, showSeconds) {
	if (!isValidDate(d)) return ""

	let H = d.getHours(),
		m = d.getMinutes(),
		s = d.getSeconds(),
		strH = String(H).padStart(2, "0"),
		strM = String(m).padStart(2, "0"),
		strS = String(s).padStart(2, "0")

	if (typeof showSeconds == "boolean" && showSeconds) {
		return `${strH}:${strM}:${strS}`
	} else {
		m = s >= 30 ? m + 1 : m
		strM = String(m).padStart(2, "0")
		return `${strH}:${strM}`
	}
}

export function isoDateString(d) {
	if (!isValidDate(d)) return ""
	let date = String(d.getDate()).padStart(2, "0"),
		month = String(d.getMonth() + 1).padStart(2, "0"),
		year = d.getFullYear()

	return `${year}-${month}-${date}`
}

export function dayName(day) {
	switch (day) {
		case 0: return "Minggu"
		case 1: return "Senin"
		case 2: return "Selasa"
		case 3: return "Rabu"
		case 4: return "Kamis"
		case 5: return "Jum'at"
		case 6: return "Sabtu"
	}
}

export function fullDate(d) {
	if (!isValidDate(d)) return ""
	return new Intl.DateTimeFormat("id", {
		day: "2-digit",
		month: "long",
		year: "numeric",
	}).format(d).replace(/\s+M$/i, "")
}

export function hijriDate(d) {
	if (!isValidDate(d)) return ""
	return new Intl.DateTimeFormat("id-u-ca-islamic", {
		day: "2-digit",
		month: "long",
		year: "numeric",
	}).format(d).replace(/\s+H$/i, "")
}