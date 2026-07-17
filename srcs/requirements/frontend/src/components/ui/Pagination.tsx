import {ReactNode} from "react";
import {LeftIcon, RightIcon} from "@/components/Icons";
import IconButton from "@/components/ui/Button/IconButton";

export default function Pagination({children, currenIndex, totalPage, onClick, variableMT=false} : {children: ReactNode, currenIndex: number, totalPage: number, onClick: (i: number) => void, variableMT?: boolean}) {
    const handleLeftArrow = () => {
        const index = currenIndex - 1;

        if (index >= 0)
            onClick(index);
    }

    const handleRightArrow = () => {
        const index = currenIndex + 1;

        if (index < totalPage)
            onClick(index);
    }

    return (<div className="w-full">
        {children}
        {totalPage > 1 &&
            <div className={"flex w-full gap-2 justify-center " + (variableMT ? "mt-2 md:mt-4 xl:mt-6" : "mt4")}>
                <IconButton disabled={currenIndex === 0} className="mt-1" onClick={handleLeftArrow} color={"gray"}>
                    {(color: string) => <LeftIcon color={color}/>}
                </IconButton>

                {Array.from({length: totalPage}, (_, i) => (
                    <button key={i} className={"custom-condensed text-2xl leading-6 " + (i === currenIndex ? "text-black font-bold" : "text-gray hover:underline")} onClick={() => {onClick(i)}}>
                        {i}
                    </button>
                ))}

                <IconButton disabled={currenIndex + 1 === totalPage} className="mt-1" onClick={handleRightArrow} color={"gray"}>
                    {(color: string) => <RightIcon color={color}/>}
                </IconButton>
        </div>}
    </div>);
}
