"use client";

export default function ModalLayout({ children, onClose, }: { children: React.ReactNode; onClose: () => void; }) {
    return (
        <div
            onClick={onClose}
            style={{
                position: "fixed",
                inset: 0,
                background:
                    "rgba(0,0,0,0.6)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                zIndex: 1000,
            }}
        >
            <div
                onClick={(e) =>
                    e.stopPropagation()
                }
                style={{
                    background: "white",
                    padding: "2rem",
                    borderRadius: "8px",
                    width: "300px",
                }}
            >
                {children}
            </div>
        </div>
    );
}
