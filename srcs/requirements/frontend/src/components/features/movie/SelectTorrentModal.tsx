"use client";

import React, {useMemo, useState} from "react";
import useModal from "@/contexts/ModalContext";
import ModalLayout from "@/components/layout/ModalLayout";
import {useTranslations} from "next-intl";
import {iTorrent} from "@/types/movie";
import Button from "@/components/ui/Button/Button";
import Pagination from "@/components/ui/Pagination";
import {SortIcon} from "@/components/Icons";

type SortKey = "source" | "quality" | "size" | "language" | "seeds";
type SortDir = "asc" | "desc";

export default function SelectTorrentModal() {
    const itemPerPage = 10;
    const {activeModal, closeModal} = useModal();
    const t = useTranslations("modal.selectTorrent");

    const [index, setIndex] = useState(0);
    const [sortKey, setSortKey] = useState<SortKey>("seeds");
    const [sortDir, setSortDir] = useState<SortDir>("desc");

    const torrents = useMemo(() => {
        return activeModal.torrents ?? [];
    }, [activeModal.torrents]);

    const totalPage = useMemo(() => {
        return Math.ceil(torrents.length / itemPerPage);
    }, [torrents.length]);

    const sortedTorrents = useMemo(() => {
        return [...torrents].sort((a, b) => {
            let valA: string | number = a[sortKey];
            let valB: string | number = b[sortKey];

            if (typeof valA === "string")
                valA = valA.toLowerCase();
            if (typeof valB === "string")
                valB = valB.toLowerCase();

            if (sortKey === "seeds") {
                valA = Number(valA);
                valB = Number(valB);
            }

            if (valA < valB)
                return sortDir === "asc" ? -1 : 1;
            if (valA > valB)
                return sortDir === "asc" ? 1 : -1;
            return 0;
        });
    }, [torrents, sortKey, sortDir]);

    const changeSort = (key: SortKey) => {
        if (key === sortKey)
            setSortDir(sortDir === "asc" ? "desc" : "asc");
        else {
            setSortKey(key);
            setSortDir("desc");
        }
        setIndex(0);
    };

    if (activeModal.type !== "select-torrent" || !activeModal.torrents || !activeModal.setTorrentId)
        return null;

    const renderHeader = (key: SortKey, label: string, hidden=false) => (
        <th onClick={() => changeSort(key)} className={"cursor-pointer select-none p-1 sm:px-3 sm:py-2 text-left text-xs sm:text-sm" + (hidden ? " hidden sm:table-cell" : "")}>
            <div className="flex items-center gap-1">
                {label}
                {sortKey === key && <SortIcon sideUp={sortDir === "asc"} />}
            </div>
        </th>);

    return (<ModalLayout onCloseAction={closeModal} title={t("title")} addMaxWTitle={false}>
        <Pagination currenIndex={index} totalPage={totalPage} onClick={setIndex}>
            <table className="w-full text-sm">
                <thead>
                <tr>
                    {renderHeader("source", t("columns.source"), true)}
                    {renderHeader("quality", t("columns.quality"))}
                    {renderHeader("size", t("columns.size"))}
                    {renderHeader("language", t("columns.language"))}
                    {renderHeader("seeds", t("columns.seeds"))}
                    <th />
                </tr>
                </thead>

                <tbody>
                {sortedTorrents.slice(index * itemPerPage, index * itemPerPage + itemPerPage).map((torrent) => (
                    <TorrentRow key={torrent.id} torrent={torrent} setTorrentId={activeModal.setTorrentId} closeModal={closeModal} t={t}/>
                ))}
                </tbody>
            </table>
        </Pagination>
    </ModalLayout>);
}

function TorrentRow({torrent, setTorrentId, closeModal, t}: {torrent: iTorrent; setTorrentId?: (id: string) => void; closeModal: () => void; t: (key: string) => string}) {
    const className = "p-1 sm:px-3 sm:py-2 text-xs sm:text-sm text-nowrap";

    return (<tr className="border-t">
        <td className={className + " text-center hidden sm:table-cell"}>{torrent.source}</td>
        <td className={className + " text-center"}>{torrent.quality}</td>
        <td className={className + " text-right"}>{`${Math.round(torrent.size)} ${t("gb")}`}</td>
        <td className={className + " text-right"}>{torrent.language}</td>
        <td className={className + " text-right"}>{torrent.seeds}</td>
        <td className={className}>
            <Button onClick={() => {
                    closeModal();
                    if (setTorrentId)
                        setTorrentId(torrent.id);
                }}>{t("choose")}</Button>
        </td>
    </tr>);
}