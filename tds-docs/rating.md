---
title: "Rating - Toss Design System | React Native"
source: "https://tossmini-docs.toss.im/tds-react-native/components/rating/"
---

# Rating - Toss Design System | React Native

> 원본: https://tossmini-docs.toss.im/tds-react-native/components/rating/

## Rating

Rating 컴포넌트는 점수를 표시하거나 사용자의 입력을 받을 수 있어요. 주로 콘텐츠에 대한 평가를 보여주거나 평가를 진행하기 위해 사용돼요.

### 사용자와 상호작용하기

Rating 컴포넌트는 두 가지 방식으로 사용자에게 정보를 제공해요:

- 읽기 전용: 사용자는 컴포넌트로부터 정보를 확인할 수 있지만, 컴포넌트와 상호작용할 수 없어요.
- 상호 작용: 사용자가 컴포넌트를 직접 제어할 수 있어요. 주로 사용자 입력을 받기 위한 용도로 사용돼요.

Rating 컴포넌트를 읽기 전용 모드로 사용하려면 readonly 속성을 설정하세요. readonly를 true로 설정하면 읽기 전용 모드로 사용할 수 있어요. 반대로, readonly를 false로 설정하면 사용자가 컴포넌트와 상호작용할 수 있어요.

### 읽기 전용

읽기 전용 모드는 readonly 속성을 true로 설정하여 사용할 수 있어요. 사용자는 컴포넌트를 클릭하거나 터치하여 상호작용할 수 없어요.

#### 크기 조정하기

Rating 컴포넌트의 크기를 변경하려면 size 속성을 사용하세요. tiny, small, medium, large, big 중 하나를 선택할 수 있어요.

```
<Rating readonly value={5} size="tiny" variant="full" />
<Rating readonly value={5} size="small" variant="full" />
<Rating readonly value={5} size="medium" variant="full" />
<Rating readonly value={5} size="large" variant="full" />
<Rating readonly value={5} size="big" variant="full" />
```

#### 형태 변경하기

Rating 컴포넌트의 형태를 바꾸려면 variant를 사용하세요. full, compact, iconOnly 중에서 선택할 수 있어요.

```
<Rating readonly value={5} size="medium" variant="full" />
<Rating readonly value={5} size="medium" variant="compact" />
<Rating readonly value={5} size="medium" variant="iconOnly" />
```

### 상호 작용하기

상호 작용 모드는 readonly 속성을 false로 설정하여 사용할 수 있어요. 사용자는 컴포넌트를 클릭하거나 터치하여 상호작용할 수 있어요.

사용자의 입력을 받기 위해서 value와 onValueChange 속성을 함께 사용하세요. value 속성은 사용자가 선택한 점수를 나타내고, onValueChange 속성은 사용자가 점수를 선택했을 때 호출되는 콜백 함수에요.

```tsx
function EditableRating() {
  const [value, setValue] = useState(5);
  return <Rating readOnly={false} value={value} max={5} size="medium" onValueChange={setValue} />;
}
```

#### 크기 조정하기

Rating 컴포넌트의 크기를 변경하려면 size 속성을 사용하세요. medium, large, big 중 하나를 선택할 수 있어요.

```
<Rating readonly={false} value={valueMedium} size="medium" onValueChange={setValueMedium} />
<Rating readonly={false} value={valueLarge} size="large" onValueChange={setValueLarge} />
<Rating readonly={false} value={valueBig} size="big" onValueChange={setValueBig} />
```

#### 형태 변경하기

상호 작용하기 모드에서는 형태를 변경할 수 없어요. variant 속성은 읽기 전용 모드에서만 사용할 수 있어요.

#### 비활성화하기

Rating 컴포넌트를 비활성화하려면 disabled 속성을 사용하세요. 사용자와 상호작용할 수 없고, 시각적으로도 비활성화된 상태임을 나타내요.

```
<Rating readonly={false} value={value} size="medium" disabled onValueChange={handleValueChange} />
<Rating readonly={false} value={value} size="large" disabled onValueChange={handleValueChange} />
<Rating readonly={false} value={value} size="big" disabled onValueChange={handleValueChange} />
```

## 인터페이스

Rating 컴포넌트는 readonly 속성에 따라 EditableRating 와 ReadOnlyRating 컴포넌트로 분기돼요.

- readonly 속성이 false이면 EditableRatingProps를 확인하세요.
- readonly 속성이 true이면 ReadOnlyRatingProps를 확인하세요.

#### EditableRatingProps

View 컴포넌트를 확장하여 제작했어요. View 컴포넌트의 모든 속성을 사용할 수 있어요.

| 속성 | 기본값 | 타입 |
| --- | --- | --- |
| readOnly* | false | false 이 값이 `false`일 때 `Rating` 컴포넌트를 제어할 수 있어요. |
| value* | - | number `Rating` 컴포넌트의 현재 점수를 결정해요. |
| size* | - | "medium" | "large" | "big" `Rating` 컴포넌트의 크기를 결정해요. |
| onValueChange | undefined | (value: number) => void `Rating` 컴포넌트의 점수 상태가 바뀔 때 실행되는 함수에요. |
| max | 5 | number `Rating` 컴포넌트에 지정 가능한 최대 점수를 결정해요. |
| gap | - | number `Rating` 컴포넌트의 요소 간의 간격을 결정해요. 지정하지 않으면 `size`에 따라 미리 정의된 값이 할당돼요. |
| activeColor | adaptive.yellow400 | string `Rating` 컴포넌트를 클릭하거나 드래그할 때, 선택된 아이콘의 색상이 activeColor로 변경돼요. 활성 상태에서는 지정한 값으로 표현돼요. 비활성 상태에서는 `adaptive.greyOpacity200`로 표현돼요. |
| disabled | false | false | true 이 값이 `true` 일 때 `Rating` 컴포넌트가 비활성화돼요. |

#### ReadOnlyRatingProps

View 컴포넌트를 확장하여 제작했어요. View 컴포넌트의 모든 속성을 사용할 수 있어요.

| 속성 | 기본값 | 타입 |
| --- | --- | --- |
| readOnly* | - | true 이 값이 `true`일 때 `Rating` 컴포넌트를 제어할 수 없어요. |
| value* | - | number `Rating` 컴포넌트의 현재 점수를 결정해요. |
| variant* | - | "full" | "compact" | "iconOnly" `Rating` 컴포넌트의 형태를 결정해요. - `full`: 전체 아이콘과 점수가 함께 보여져요. - `compact`: 하나의 아이콘과 점수가 함께 보여져요. - `iconOnly`: 전체 아이콘만 보여져요. |
| size* | - | "medium" | "large" | "big" | "tiny" | "small" `Rating` 컴포넌트의 크기를 결정해요. |
| max | 5 | number `Rating` 컴포넌트에 지정 가능한 최대 점수를 결정해요. |
| gap | - | number `Rating` 컴포넌트의 요소 간의 간격을 결정해요. 지정하지 않으면 `size`에 따라 미리 정의된 값이 할당돼요. |
| activeColor | adaptive.yellow400 | string `Rating` 컴포넌트의 활성 색상을 지정해요. 비활성 상태에서는 `adaptive.greyOpacity200`로 표현돼요. |
