"use client";

import React, {createContext, Dispatch, SetStateAction, useContext, useState} from "react";

type ModalType = | "signin" | "register" | "genre" | "filter-genre" | "forgot-password" | "delete-comment" | null;

interface ModalState {
    type: ModalType;
    genres?: string[];
    filterGenre?: [filterGenre: string[], handleFilterGenre: (newGenres: string[]) => void];
    setFilterGenre?: Dispatch<SetStateAction<string[]>>
    commentId?: number
    deleteComment?: (commentId: number) => void
}

interface ModalContextType {
    activeModal: ModalState;
    openModal: (modal: ModalState) => void;
    closeModal: () => void;
}

const ModalContext = createContext<ModalContextType | null>(null);

export function ModalProvider({ children, }: { children: React.ReactNode; }) {
    const [activeModal, setActiveModal] = useState<ModalState>({type: null});

    const openModal = (modal: ModalState) => setActiveModal(modal);
    const closeModal = () => setActiveModal({type: null});

    return (<ModalContext.Provider value={{activeModal, openModal, closeModal,}}>
        {children}
    </ModalContext.Provider>);
}

export function useModal() {
    const context = useContext(ModalContext);
    if (!context)
        throw new Error("useModal must be used inside ModalProvider");
    return context;
}
