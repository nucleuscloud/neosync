'use client';

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import * as AccordionPrimitive from '@radix-ui/react-accordion';
import { ChevronRightIcon } from '@radix-ui/react-icons';
import * as React from 'react';
import useResizeObserver from 'use-resize-observer';
import { Checkbox } from './checkbox';
import { ScrollArea } from './scroll-area';

interface TreeDataItem {
  id: string;
  name: string;
  isSelected?: boolean;
  children?: TreeDataItem[];
}

type TreeProps = React.HTMLAttributes<HTMLDivElement> & {
  data: TreeDataItem[];
  onSelectChange?: (items: TreeDataItem[]) => void;
};

const Tree = React.forwardRef<HTMLDivElement, TreeProps>(
  ({ data, onSelectChange, className, ...props }, ref) => {
    const [treeItems, setTreeItems] = React.useState<TreeDataItem[]>(data);

    React.useEffect(() => {
      setTreeItems(data);
    }, [data]);

    function handleSelectChange(item: TreeDataItem) {
      const newTree = updateItemAndChildren(
        treeItems,
        item.id,
        !item.isSelected
      );
      setTreeItems(newTree);
      if (onSelectChange) {
        onSelectChange(newTree);
      }
    }

    function updateItemAndChildren(
      items: TreeDataItem[],
      id: string,
      isSelected: boolean,
      foundItem: boolean = false
    ): TreeDataItem[] {
      return items.map((item) => {
        const isCurrentOrFound = item.id === id || foundItem;

        let updatedChildren = item.children;
        if (item.children) {
          updatedChildren = updateItemAndChildren(
            item.children,
            id,
            isSelected,
            isCurrentOrFound
          );
        }

        let updatedIsSelected = item.isSelected;
        if (isCurrentOrFound) {
          updatedIsSelected = isSelected;
        } else if (item.children) {
          updatedIsSelected = updatedChildren?.some(
            (child) => child.isSelected
          );
        }

        return {
          ...item,
          isSelected: updatedIsSelected,
          children: updatedChildren,
        };
      });
    }

    function isIndeterminate(item: TreeDataItem): boolean {
      if (!item.children) {
        return false;
      }
      function walkTreeItems(items: TreeDataItem | TreeDataItem[]) {
        if (items instanceof Array) {
          // eslint-disable-next-line @typescript-eslint/prefer-for-of
          for (let i = 0; i < items.length; i++) {
            if (!items[i].isSelected) {
              return true;
            }
            if (walkTreeItems(items[i])) {
              return true;
            }
          }
        } else if (items.children) {
          return walkTreeItems(items.children);
        }
        return false;
      }

      return walkTreeItems(item);
    }

    const { ref: refRoot, width, height } = useResizeObserver();

    return (
      <div ref={refRoot} className={cn('p-2', className)}>
        <ScrollArea style={{ width, height }}>
          <TreeItem
            data={treeItems}
            ref={ref}
            handleSelectChange={handleSelectChange}
            isIndeterminate={isIndeterminate}
            {...props}
          />
        </ScrollArea>
      </div>
    );
  }
);
Tree.displayName = 'Tree';

type TreeItemProps = React.HTMLAttributes<HTMLDivElement> & {
  data: TreeDataItem | TreeDataItem[];
  onSelectChange?: (items: TreeDataItem[]) => void;
  handleSelectChange: (item: TreeDataItem) => void;
  isIndeterminate: (item: TreeDataItem) => boolean;
};

const TreeItem = React.forwardRef<HTMLDivElement, TreeItemProps>(
  ({ className, data, handleSelectChange, isIndeterminate, ...props }, ref) => {
    return (
      <div ref={ref} role="tree" className={className} {...props}>
        <ul>
          {data instanceof Array ? (
            data.map((item) => (
              <li key={item.id}>
                {item.children ? (
                  <AccordionPrimitive.Root
                    type="multiple"
                    defaultValue={[item.id]}
                  >
                    <AccordionPrimitive.Item value={item.id}>
                      <div className="flex flex-row justify-between ">
                        <div className="flex items-center space-x-2">
                          <Checkbox
                            id={item.id}
                            onClick={() => handleSelectChange(item)}
                            checked={item.isSelected}
                            indeterminate={isIndeterminate(item)}
                          />
                          <label
                            htmlFor={item.id}
                            className="text-sm truncate font-medium"
                          >
                            {item.name}
                          </label>
                        </div>
                        <AccordionTrigger
                          className={cn(
                            'px-2 hover:before:opacity-100 before:absolute before:left-0 before:w-full before:opacity-0 before:bg-muted/80 before:h-[1.75rem] before:-z-10'
                          )}
                        ></AccordionTrigger>
                      </div>
                      <AccordionContent className="pl-6">
                        <TreeItem
                          data={item.children ? item.children : item}
                          handleSelectChange={handleSelectChange}
                          isIndeterminate={isIndeterminate}
                        />
                      </AccordionContent>
                    </AccordionPrimitive.Item>
                  </AccordionPrimitive.Root>
                ) : (
                  <Leaf
                    item={item}
                    handleSelectChange={() => handleSelectChange(item)}
                    indeterminate={isIndeterminate(item)}
                  />
                )}
              </li>
            ))
          ) : (
            <li>
              <Leaf
                item={data}
                handleSelectChange={() => handleSelectChange(data)}
                indeterminate={isIndeterminate(data)}
              />
            </li>
          )}
        </ul>
      </div>
    );
  }
);
TreeItem.displayName = 'TreeItem';

const Leaf = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    item: TreeDataItem;
    handleSelectChange: (item: TreeDataItem | undefined) => void;
    indeterminate: boolean;
  }
>(({ className, handleSelectChange, indeterminate, item, ...props }, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        'flex items-center py-2 \
        hover:before:opacity-100 before:absolute before:left-0 before:right-1 before:w-full before:opacity-0 before:bg-muted/80 before:h-[1.75rem] before:-z-10',
        className
      )}
      {...props}
    >
      <div className={`flex items-center space-x-2`}>
        <Checkbox
          id={item.id}
          onClick={() => handleSelectChange(item)}
          checked={item.isSelected}
          indeterminate={indeterminate}
        />
        <div className="flex ">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <label htmlFor={item.id} className={`text-sm font-medium`}>
                  {item.name}
                </label>
              </TooltipTrigger>
              <TooltipContent>{item.name}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
    </div>
  );
});
Leaf.displayName = 'Leaf';

const AccordionTrigger = React.forwardRef<
  React.ElementRef<typeof AccordionPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
  <AccordionPrimitive.Header>
    <AccordionPrimitive.Trigger
      ref={ref}
      className={cn(
        'flex flex-1 w-full items-center py-2 transition-all last:[&[data-state=open]>svg]:rotate-90',
        className
      )}
      {...props}
    >
      {children}
      <ChevronRightIcon className="h-4 w-4 shrink-0 transition-transform duration-200 text-accent-foreground/50 ml-auto" />
    </AccordionPrimitive.Trigger>
  </AccordionPrimitive.Header>
));
AccordionTrigger.displayName = AccordionPrimitive.Trigger.displayName;

const AccordionContent = React.forwardRef<
  React.ElementRef<typeof AccordionPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <AccordionPrimitive.Content
    ref={ref}
    className={cn(
      'overflow-hidden text-sm transition-all data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down',
      className
    )}
    {...props}
  >
    <div className="pb-1 pt-0">{children}</div>
  </AccordionPrimitive.Content>
));
AccordionContent.displayName = AccordionPrimitive.Content.displayName;

export { Tree, type TreeDataItem };
