'use client';

import { cn } from '@/libs/utils';
import * as AccordionPrimitive from '@radix-ui/react-accordion';
import { ChevronRight, type LucideIcon } from 'lucide-react';
import * as React from 'react';
import useResizeObserver from 'use-resize-observer';
import { Checkbox } from './checkbox';
import { ScrollArea } from './scroll-area';

interface TreeDataItem {
  id: string;
  name: string;
  isSelected?: boolean;
  icon?: LucideIcon;
  children?: TreeDataItem[];
}

type TreeProps = React.HTMLAttributes<HTMLDivElement> & {
  data: TreeDataItem[];
  onSelectChange?: (items: TreeDataItem[]) => void;
  folderIcon?: LucideIcon;
  itemIcon?: LucideIcon;
};

const Tree = React.forwardRef<HTMLDivElement, TreeProps>(
  (
    { data, onSelectChange, folderIcon, itemIcon, className, ...props },
    ref
  ) => {
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

        if (item.children) {
          return {
            ...item,
            isSelected: isCurrentOrFound ? isSelected : item.isSelected,
            children: updateItemAndChildren(
              item.children,
              id,
              isSelected,
              isCurrentOrFound
            ),
          };
        }

        if (isCurrentOrFound) {
          return { ...item, isSelected };
        }

        return item;
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
            console.log('item', JSON.stringify(items[i]));
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
      <div ref={refRoot} className={cn('overflow-hidden ', className)}>
        <ScrollArea style={{ width, height }}>
          <div className="relative p-2">
            <TreeItem
              data={treeItems}
              ref={ref}
              handleSelectChange={handleSelectChange}
              FolderIcon={folderIcon}
              ItemIcon={itemIcon}
              isIndeterminate={isIndeterminate}
              {...props}
            />
          </div>
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
  FolderIcon?: LucideIcon;
  ItemIcon?: LucideIcon;
};

const TreeItem = React.forwardRef<HTMLDivElement, TreeItemProps>(
  (
    {
      className,
      data,
      handleSelectChange,
      isIndeterminate,
      FolderIcon,
      ItemIcon,
      ...props
    },
    ref
  ) => {
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
                            htmlFor="terms"
                            className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                          >
                            {item.name}
                          </label>
                        </div>
                        <AccordionTrigger
                          className={cn(
                            '  px-2 hover:before:opacity-100 before:absolute before:left-0 before:w-full before:opacity-0 before:bg-muted/80 before:h-[1.75rem] before:-z-10'
                          )}
                        >
                          {item.icon && (
                            <item.icon
                              className="h-4 w-4 shrink-0 mr-2 text-accent-foreground/50"
                              aria-hidden="true"
                            />
                          )}
                          {!item.icon && FolderIcon && (
                            <FolderIcon
                              className="h-4 w-4 shrink-0 mr-2 text-accent-foreground/50"
                              aria-hidden="true"
                            />
                          )}
                        </AccordionTrigger>
                      </div>
                      <AccordionContent className="pl-6">
                        <TreeItem
                          data={item.children ? item.children : item}
                          handleSelectChange={handleSelectChange}
                          FolderIcon={FolderIcon}
                          ItemIcon={ItemIcon}
                          isIndeterminate={isIndeterminate}
                        />
                      </AccordionContent>
                    </AccordionPrimitive.Item>
                  </AccordionPrimitive.Root>
                ) : (
                  <Leaf
                    item={item}
                    handleSelectChange={() => handleSelectChange(item)}
                    Icon={ItemIcon}
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
                Icon={ItemIcon}
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
    Icon?: LucideIcon;
    handleSelectChange: (item: TreeDataItem | undefined) => void;
    indeterminate: boolean;
  }
>(
  (
    { className, handleSelectChange, indeterminate, item, Icon, ...props },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center py-2 px-2 cursor-pointer \
        hover:before:opacity-100 before:absolute before:left-0 before:right-1 before:w-full before:opacity-0 before:bg-muted/80 before:h-[1.75rem] before:-z-10',
          className
        )}
        {...props}
      >
        {item.icon && (
          <item.icon
            className="h-4 w-4 shrink-0 mr-2 text-accent-foreground/50"
            aria-hidden="true"
          />
        )}
        {!item.icon && Icon && (
          <Icon
            className="h-4 w-4 shrink-0 mr-2 text-accent-foreground/50"
            aria-hidden="true"
          />
        )}
        <div className="flex items-center space-x-2">
          <Checkbox
            id={item.id}
            onClick={() => handleSelectChange(item)}
            checked={item.isSelected}
            indeterminate={indeterminate}
          />
          <label
            htmlFor="terms"
            className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            {item.name}
          </label>
        </div>
      </div>
    );
  }
);
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
      <ChevronRight className="h-4 w-4 shrink-0 transition-transform duration-200 text-accent-foreground/50 ml-auto" />
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
