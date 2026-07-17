export default function Colors({className, heigth="h-3 sm:h-4"}: {className?: string, heigth?: string}) {
    return (<div className={"w-full relative " + className + " " + heigth}>
        <div className="custom-noise" />
        <div className="flex h-full">
            <div className="size-full bg-yellow hover:bg-yellow-hover" />
            <div className="size-full bg-pink hover:bg-pink-hover" />
            <div className="size-full bg-green hover:bg-green-hover" />
            <div className="size-full bg-purple hover:bg-purple-hover" />
            <div className="size-full bg-blue hover:bg-blue-hover" />
            <div className="size-full bg-red hover:bg-red-hover" />
        </div>
    </div>);
}
