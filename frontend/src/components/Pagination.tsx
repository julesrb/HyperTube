import {ReactNode, useState} from "react";
import {LeftIcon, RightIcon} from "@/components/Icon";

export default function Pagination({children, currenIndex, totalPage, onClick} : {children: ReactNode, currenIndex: number, totalPage: number, onClick: (i: number) => void}) {
    const [leftColor, setLeftColor] = useState("gray");
    const [rightColor, setRightColor] = useState("gray");

    const handleLeftArrow = () => {
        const index = currenIndex - 1;

        if (index >= 0) {
            onClick(index);
        }
    }

    const handleRightArrow = () => {
        const index = currenIndex + 1;

        if (index < totalPage) {
            onClick(index);
        }
    }

    return (<div>
        {children}
        <div className="flex w-full gap-2 justify-center my-4">
            <button className="mt-1" onClick={handleLeftArrow} onMouseEnter={() => setLeftColor(currenIndex === 0 ? "gray" : "black")} onMouseLeave={() => setLeftColor("gray")}>
                <LeftIcon color={leftColor}/>
            </button>
            {Array.from({ length: totalPage }, (_, i) => (
                <button key={i} className={"custom-condensed text-2xl leading-6 " + (i === currenIndex ? "text-black font-bold" : "text-gray hover:underline")} onClick={() => {onClick(i)}}>
                    {i}
                </button>
            ))}

            <button className="mt-1" onClick={handleRightArrow} onMouseEnter={() => setRightColor(currenIndex + 1 === totalPage ? "gray" : "black")} onMouseLeave={() => setRightColor("gray")}>
                <RightIcon color={rightColor}/>
            </button>
        </div>
    </div>);
}