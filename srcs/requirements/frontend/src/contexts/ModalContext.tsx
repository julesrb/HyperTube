"use client";

import React, {createContext, Dispatch, SetStateAction, useContext, useEffect, useState} from "react";
import {iGenre} from "@/types/genre";
import {iTorrent} from "@/types/movie";

type ModalType = "signin" | "register" | "genre" | "filter-genre" | "send-email-forgot-password" | "set-new-password" | "delete-confirmation" | "select-torrent" | "application" | null;

export interface ModalState {
    type: ModalType;
    genres?: number[];
    filterGenre?: [filterGenre: iGenre[], handleFilterGenre: (newGenres: iGenre[]) => void];
    setFilterGenre?: Dispatch<SetStateAction<iGenre[]>>
    deleteObjId?: number
    deleteFunc?: (objId: number) => Promise<void>;
    token?: string
    torrents?: iTorrent[]
    setTorrentId?: (selectTorrentId: string) => void
    noClose?: boolean
    reload?: () => void
    appId?: number
    pageIndex?: number
}

interface ModalContextType {
    activeModal: ModalState;
    openModal: (modal: ModalState) => void;
    closeModal: () => void;
}

const ModalContext = createContext<ModalContextType | null>(null);

export function ModalProvider({children}: {children: React.ReactNode}) {
    const [activeModal, setActiveModal] = useState<ModalState>({type: null});

    const openModal = (modal: ModalState) => setActiveModal(modal);
    const closeModal = () => setActiveModal({type: null});

    useEffect(() => {
        if (activeModal.type !== null) {
            document.body.style.overflow = "hidden";
        } else {
            document.body.style.overflow = "";
        }

        return () => {document.body.style.overflow = "";};
    }, [activeModal]);

    return (<ModalContext.Provider value={{activeModal, openModal, closeModal}}>
        {children}
    </ModalContext.Provider>);
}

export default function useModal() {
    const context = useContext(ModalContext);
    if (!context)
        throw new Error("useModal must be used inside ModalProvider");
    return context;
}
