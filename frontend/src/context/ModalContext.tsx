"use client";

import React, { createContext, useContext, useState } from "react";

type ModalType =
    | "signin"
    | "register"
    | "genre"
    | null;

type ModalData = | string[] | undefined;

interface ModalState {
    type: ModalType;
    data?: ModalData;
}

interface ModalContextType {
    activeModal: ModalState;
    openModal: (type: ModalType, data?: ModalData) => void;
    closeModal: () => void;
}

const ModalContext = createContext<ModalContextType | null>(null);

export function ModalProvider({ children, }: { children: React.ReactNode; }) {
    const [activeModal, setActiveModal] = useState<ModalState>({type: null, data: undefined});

    const openModal = (type: ModalType, data?: ModalData) => setActiveModal({type, data});
    const closeModal = () => setActiveModal({type: null, data: undefined});

    return (
        <ModalContext.Provider value={{activeModal, openModal, closeModal,}}>
            {children}
        </ModalContext.Provider>
    );
}

export function useModal() {
    const context = useContext(ModalContext);

    if (!context) {
        throw new Error("useModal must be used inside ModalProvider");
    }

    return context;
}
