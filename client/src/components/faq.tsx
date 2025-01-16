import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"

import TextAnimate from "./cult/text-animate"

export function FAQ() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-24 sm:py-32 lg:px-8 lg:py-20 bg-transparent rounded-t-[48px]">
      <div className="py-12">
        <TextAnimate
          text="Frequently asked "
          type="shiftInUp"
          className="md:text-[6rem] text-[2rem] font-bold md:leading-10 md:pb-14 tracking-tight text-orange-100 md:py-8"
        />
        <TextAnimate
          text="questions. "
          type="shiftInUp"
          className="md:text-[6rem] text-[2rem] font-bold md:leading-10 tracking-tight text-orange-100 md:py-8 font-brand"
        />
      </div>

      {/* CREDIT BG PATTERN -  https://bg.ibelick.com/ */}
      <div className="absolute inset-0 -z-10 h-full w-full items-center px-5 py-24 [background:radial-gradient(125%_125%_at_50%_10%,#000_50%,#FF7C33_100%)]"></div>

      <div className="md:mx-auto">
        <Accordion
          type="multiple"
          className="w-full md:space-y-9 bg-black/10 rounded-xl border border-orange-50/20 text-white  backdrop-blur "
        >
          <AccordionItem
            value="item-1"
            className="border-x border-b-0 border-black/10 rounded-md md:px-4"
          >
            <AccordionTrigger className=" text-xl md:text-3xl text-left pr-4 md:pr-0  font-medium">
              <span className="px-6 md:px-2">
                Why not just hire a full-time design engineer?
              </span>
            </AccordionTrigger>
            <AccordionContent className="text-lg font-semibold text-neutral-300 px-4">
              Good luck! The annual cost of a full-time design engineer capable
              of design, front-end development, backend development, and
              database design exceeds $200,000, plus benefits. Thats before you
              price in recruiting costs, interview time etc. Whats great about a
              subscription is low risk with the highest quality outcome. Its a
              rare win win :)
            </AccordionContent>
          </AccordionItem>
          <AccordionItem
            value="item-2"
            className="border-x border-b-0 border-black/10 rounded-md px-4"
          >
            <AccordionTrigger className=" text-xl md:text-3xl text-left pl-2  font-medium">
              Is there a limit to how many requests I can have?
            </AccordionTrigger>
            <AccordionContent className="text-lg font-semibold text-neutral-300 pl-2">
              Once subscribed, you're able to add as many feature requests to
              your queue as you'd like, and they will be delivered one by one..
              unless you're in goblin mode.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem
            value="item-3"
            className="border-x border-b-0 border-black/10 rounded-md px-4"
          >
            <AccordionTrigger className=" text-xl md:text-3xl text-left pl-2  font-medium">
              How long will it take to build a full stack feature?
            </AccordionTrigger>
            <AccordionContent className="text-lg font-semibold text-neutral-300 pl-2">
              Most features are completed in just two weeks or less. However,
              more complex features can take longer.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem
            value="item-4"
            className="border-x border-b-0 border-black/10 rounded-md px-4"
          >
            <AccordionTrigger className=" text-xl md:text-3xl text-left pl-2  font-medium">
              Who is the team?
            </AccordionTrigger>
            <AccordionContent className="text-lg font-semibold text-neutral-300 pl-2">
              Yours truly @nolansym
            </AccordionContent>
          </AccordionItem>
          <AccordionItem
            value="item-5"
            className="border-x border-b-0 border-black/10 rounded-md px-4"
          >
            <AccordionTrigger className=" text-xl md:text-3xl text-left pl-2  font-medium">
              What if I want a different tech stack?
            </AccordionTrigger>
            <AccordionContent className="text-lg font-semibold text-neutral-300 pl-2">
              Sorry, we have picked our tech stack because its powerful, its
              popular and it allows us to ship fast.{" "}
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  )
}
