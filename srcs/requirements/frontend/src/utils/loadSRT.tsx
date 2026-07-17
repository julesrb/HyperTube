import SrtParser from "srt-parser-2";

const parser = new SrtParser();

export default async function loadSRT(url: string) {
    const res = await fetch(url);
    const srtText = await res.text();

    return parser.fromSrt(srtText).map(sub => ({
        start: timeToSeconds(sub.startTime),
        end: timeToSeconds(sub.endTime),
        text: sub.text
    }));
}

function timeToSeconds(time: string) {
    const [h, m, s] = time.split(":");
    const [sec, ms] = s.split(",");

    return (
        parseInt(h) * 3600 +
        parseInt(m) * 60 +
        parseInt(sec) +
        parseInt(ms) / 1000
    );
}
